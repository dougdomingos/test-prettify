package coverage

import (
	"html"
	"os"
	"path"
	"sort"
	"strings"

	"html/template"

	"github.com/dougdomingos/test-prettify/internal/model"
)

// RenderOptions controls how the coverage profile is rendered into a report.
type RenderOptions struct {

	// RootModule is the module/source root used to resolve coverage paths.
	// When empty, the current working directory is used.
	RootModule string
}

// clip is a half-open rune interval [start, end) on a single source line,
// tagged with whether that region was executed (covered). A profile block that
// spans multiple lines is split into one clip per line.
type clip struct {
	start   int
	end     int
	covered bool
}

// RenderProfile reads every source file referenced by the profile and produces
// the coverage page data model, highlighting the exact block boundaries
// recorded in the profile.
func RenderProfile(profile *model.Profile, opts RenderOptions) (*model.CoverageReportData, error) {
	fileResolver, err := NewSrcResolver(opts.RootModule)
	if err != nil {
		return nil, err
	}

	filePaths := make([]string, 0, len(profile.BlocksByFile))
	for file := range profile.BlocksByFile {
		filePaths = append(filePaths, file)
	}
	sort.Strings(filePaths)

	var totalStmts, coveredStmts int
	pkgStates := make(map[string]*packageReportState)
	pkgOrder := make([]string, 0, len(filePaths))
	allFiles := make([]*model.CoverageFileReport, 0, len(filePaths))

	for _, filePath := range filePaths {
		blocks := profile.BlocksByFile[filePath]
		relPath := fileResolver.GetRelPath(filePath)

		var fileStmts, fileStmtsCovered int
		for _, b := range blocks {
			fileStmts += b.NumStmt
			if b.Count > 0 {
				fileStmtsCovered += b.NumStmt
			}
		}
		
		totalStmts += fileStmts
		coveredStmts += fileStmtsCovered

		report := &model.CoverageFileReport{Path: relPath}
		if fileStmts > 0 {
			report.Percent = float64(fileStmtsCovered) / float64(fileStmts) * 100
		}

		if abs, ok := fileResolver.Resolve(filePath); ok {
			rows, err := buildFileCoverageLines(abs, blocks)
			if err != nil {
				report.Unavailable = true
			} else {
				report.Lines = rows
			}
		} else {
			report.Unavailable = true
		}

		pkgName := getPackageName(relPath)
		acc, ok := pkgStates[pkgName]
		if !ok {
			acc = &packageReportState{name: pkgName}
			pkgStates[pkgName] = acc
			pkgOrder = append(pkgOrder, pkgName)
		}
		
		acc.total += fileStmts
		acc.files = append(acc.files, report)
		acc.covered += fileStmtsCovered
		
		allFiles = append(allFiles, report)
	}

	sort.Strings(pkgOrder)

	sort.SliceStable(allFiles, func(i, j int) bool {
		if allFiles[i].Percent != allFiles[j].Percent {
			return allFiles[i].Percent < allFiles[j].Percent
		}
		return allFiles[i].Path < allFiles[j].Path
	})

	lowest := make([]model.CoverageFileSummary, 0, len(allFiles))
	for i, report := range allFiles {
		if i == 3 {
			break
		}
		
		lowest = append(lowest, model.CoverageFileSummary{
			Path:    report.Path,
			Percent: report.Percent,
		})
	}

	data := &model.CoverageReportData{
		Overview: model.CoverageOverview{
			CoveredStmts: coveredStmts,
			TotalStmts:   totalStmts,
			FileCount:    len(filePaths),
			PackageCount: len(pkgOrder),
			LowestRate: lowest,
		},
	}
	if totalStmts > 0 {
		data.Overview.GlobalPercent = float64(coveredStmts) / float64(totalStmts) * 100
	}

	for _, name := range pkgOrder {
		acc := pkgStates[name]
		pkg := &model.CoveragePackageReport{Name: acc.name, Files: acc.files}
		if acc.total > 0 {
			pkg.Percent = float64(acc.covered) / float64(acc.total) * 100
		}
		data.Packages = append(data.Packages, pkg)
	}

	return data, nil
}

// packageReportState accumulates statement counts and file references for a
// single package while the report is being assembled.
type packageReportState struct {
	name    string
	files   []*model.CoverageFileReport
	total   int
	covered int
}

// getPackageName derives the short package name from a module-relative file
// path (e.g. "internal/auth/auth.go" -> "auth"). Files living at the module
// root are treated as package "main".
func getPackageName(rel string) string {
	dir := path.Dir(rel)
	switch dir {
	case ".", "":
		return "main"
	default:
		return path.Base(dir)
	}
}

// buildFileCoverageLines reads a source file and builds the highlighted line
// rows for the given profile blocks.
func buildFileCoverageLines(absPath string, blocks []model.Block) ([]model.CoveredLine, error) {
	raw, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(raw), "\n")
	if n := len(lines); n > 0 && lines[n-1] == "" {
		lines = lines[:n-1]
	}

	indexes := make([]*runeIndex, len(lines))
	for i, text := range lines {
		indexes[i] = newRuneIndex(text)
	}

	clipsByLine := make(map[int][]clip, len(lines))
	for _, b := range blocks {
		clipBlockToLines(indexes, b, clipsByLine)
	}

	rows := make([]model.CoveredLine, 0, len(lines))
	for idx, text := range lines {
		lineNum := idx + 1
		rows = append(rows, buildCoveredLine(lineNum, text, clipsByLine[lineNum]))
	}
	return rows, nil
}

// buildCoveredLine builds the highlighted markup and status of a single source
// line. Blank lines stay neutral regardless of block coverage to keep the
// viewer readable.
func buildCoveredLine(lineNum int, text string, clips []clip) model.CoveredLine {
	row := model.CoveredLine{LineNum: lineNum, Status: model.LineNeutral}

	if strings.TrimSpace(text) == "" {
		row.HTML = template.HTML(html.EscapeString(text))
		return row
	}

	segs := mergeClipsToSegments(text, clips)
	row.HTML = highlightLineHTML(text, segs)
	row.Status = lineStatusFromSegments(text, segs)
	return row
}

// clipBlockToLines clips a single block onto every line it spans.
func clipBlockToLines(indexes []*runeIndex, b model.Block, clipsByLine map[int][]clip) {
	numLines := len(indexes)

	startLine, endLine := b.StartLine, b.EndLine
	if startLine < 1 {
		startLine = 1
	}
	if endLine > numLines {
		endLine = numLines
	}
	if startLine > numLines || startLine > endLine {
		return
	}

	covered := b.Count > 0

	if startLine == endLine {
		idx := indexes[startLine-1]
		clipByteRange(clipsByLine, startLine, idx, b.StartCol, b.EndCol, covered)
		return
	}

	// First line: from the start column to the end of the line.
	startIdx := indexes[startLine-1]
	lo := runeAtColumn(startIdx, b.StartCol)
	if lo < startIdx.count() {
		registerClip(clipsByLine, startLine, lo, startIdx.count(), covered)
	}

	// Whole intermediate lines in between.
	for ln := startLine + 1; ln < endLine; ln++ {
		if indexes[ln-1].trimmedEmpty() {
			continue
		}
		
		registerClip(clipsByLine, ln, 0, indexes[ln-1].count(), covered)
	}

	// Last line: from the start of the line to the end column.
	endIdx := indexes[endLine-1]
	hi := runeColumnEnd(endIdx, b.EndCol)
	if hi > 0 {
		registerClip(clipsByLine, endLine, 0, hi, covered)
	}
}

// clipByteRange maps a byte column range onto a line and registers the clip.
func clipByteRange(clipsByLine map[int][]clip, lineNum int, idx *runeIndex, startCol, endCol int, covered bool) {
	start := runeAtColumn(idx, startCol)
	end := runeColumnEnd(idx, endCol)
	if end <= start {
		return
	}
	registerClip(clipsByLine, lineNum, start, end, covered)
}

// registerClip records a clip on a line, skipping empty ranges.
func registerClip(clipsByLine map[int][]clip, lineNum, start, end int, covered bool) {
	if end <= start {
		return
	}
	clipsByLine[lineNum] = append(clipsByLine[lineNum], clip{start: start, end: end, covered: covered})
}

// runeSeg is a merged, half-open rune interval [start, end) on a line that is
// either fully covered or fully uncovered. It is the result of collapsing all
// overlapping clips on a line into contiguous, non-overlapping segments.
type runeSeg struct {
	start   int
	end     int
	covered bool
}

// event marks where a clip starts (delta +1) or ends (delta -1); it is used by
// mergeClipsToSegments while sweeping a line. covered ties each change to the
// covering state of the clip so overlapping covered/uncovered regions can be
// resolved.
type event struct {
	pos     int
	delta   int
	covered bool
}

// mergeClipsToSegments merges all clips on a line into the minimal set of
// contiguous covered/uncovered segments. When a position is claimed by both a
// covered and an uncovered block, the covered state wins.
func mergeClipsToSegments(text string, clips []clip) []runeSeg {
	if len(clips) == 0 {
		return nil
	}

	events := make([]event, 0, len(clips)*2)
	for _, c := range clips {
		events = append(events, event{pos: c.start, delta: 1, covered: c.covered})
		events = append(events, event{pos: c.end, delta: -1, covered: c.covered})
	}
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].pos < events[j].pos
	})

	lineLen := runeCount(text)
	segs := make([]runeSeg, 0, len(clips))
	openCovered, openUncovered := 0, 0

	i := 0
	for i < len(events) {
		pos := events[i].pos

		j := i
		for j < len(events) && events[j].pos == pos {
			if events[j].covered {
				openCovered += events[j].delta
			} else {
				openUncovered += events[j].delta
			}
			j++
		}

		next := lineLen
		if j < len(events) {
			next = events[j].pos
		}

		if next > pos {
			switch {
			case openCovered > 0:
				segs = append(segs, runeSeg{start: pos, end: next, covered: true})
			case openUncovered > 0:
				segs = append(segs, runeSeg{start: pos, end: next, covered: false})
			}
		}

		i = j
	}

	return segs
}

// highlightLineHTML escapes the line text and wraps covered/uncovered segments
// into span tags.
func highlightLineHTML(text string, segs []runeSeg) template.HTML {
	if len(segs) == 0 {
		return template.HTML(html.EscapeString(text))
	}

	runes := []rune(text)
	var b strings.Builder
	pos := 0
	for _, s := range segs {
		if s.start > pos {
			b.WriteString(html.EscapeString(string(runes[pos:s.start])))
		}
		class := "uncov"
		if s.covered {
			class = "cov"
		}

		b.WriteString(`<span class="`);
		b.WriteString(class);
		b.WriteString(`">`)

		b.WriteString(html.EscapeString(string(runes[s.start:s.end])))
		b.WriteString(`</span>`)
		
		pos = s.end
	}

	if pos < len(runes) {
		b.WriteString(html.EscapeString(string(runes[pos:])))
	}

	return template.HTML(b.String())
}

// lineStatusFromSegments derives the line status from its segments. A line is
// only painted uniformly (green/red row) when its coverage segments account
// for every non-whitespace rune; otherwise it is Mixed and the inline spans
// carry the exact sub-line boundaries.
func lineStatusFromSegments(text string, segs []runeSeg) model.LineStatus {
	if len(segs) == 0 {
		return model.LineNeutral
	}

	runes := []rune(text)
	inSeg := make([]bool, len(runes))
	var anyCovered, anyUncovered bool
	for _, s := range segs {
		for i := s.start; i < s.end && i < len(inSeg); i++ {
			inSeg[i] = true
			if s.covered {
				anyCovered = true
			} else {
				anyUncovered = true
			}
		}
	}

	if anyCovered && anyUncovered {
		return model.LineMixed
	}

	for i, rn := range runes {
		if !inSeg[i] && !isSpace(rn) {
			return model.LineMixed
		}
	}

	if anyCovered {
		return model.LineCovered
	}
	return model.LineUncovered
}

// isSpace reports whether a rune is treated as insignificant whitespace when
// deciding if a line is fully covered.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r'
}

// runeIndex maps byte offsets to rune indices for a single source line so
// profile columns (1-based byte offsets) can be converted to rune positions
// for HTML rendering.
type runeIndex struct {
	runes  []rune
	starts []int // byte offset where each rune starts; len(starts) == len(runes)
}

// newRuneIndex builds the rune table for a line of source code.
func newRuneIndex(s string) *runeIndex {
	runes := []rune(s)
	starts := make([]int, len(runes))

	bytePos := 0
	for i, r := range runes {
		starts[i] = bytePos
		bytePos += len(string(r))
	}
	return &runeIndex{runes: runes, starts: starts}
}

// count returns the number of runes in the line.
func (r *runeIndex) count() int {
	return len(r.runes)
}

// trimmedEmpty reports whether the line holds nothing but whitespace.
func (r *runeIndex) trimmedEmpty() bool {
	for _, rn := range r.runes {
		if rn != ' ' && rn != '\t' {
			return false
		}
	}
	return true
}

// runeAtColumn maps an inclusive 1-based byte column to the 0-based rune
// index where the covered region begins.
func runeAtColumn(r *runeIndex, col int) int {
	b := col - 1
	if b <= 0 {
		return 0
	}
	if b >= r.byteLen() {
		return r.count()
	}
	return byteOffsetContainingRune(r, b)
}

// runeColumnEnd maps an exclusive 1-based byte column to the 0-based rune
// index where the covered region ends.
func runeColumnEnd(r *runeIndex, col int) int {
	b := col - 1
	if b <= 0 {
		return 0
	}
	if b >= r.byteLen() {
		return r.count()
	}

	idx := byteOffsetContainingRune(r, b)
	// The end column is exclusive: when the byte offset falls mid-rune, the
	// region ends at the next rune boundary so the final rune is kept whole.
	if idx < r.count() && b > r.starts[idx] {
		return idx + 1
	}
	
	return idx
}

// byteLen returns the total byte length of the line.
func (r *runeIndex) byteLen() int {
	if len(r.starts) == 0 {
		return 0
	}
	last := r.starts[len(r.starts)-1]
	return last + len(string(r.runes[len(r.runes)-1]))
}

// byteOffsetContainingRune returns the index of the rune that contains the
// given 0-based byte offset.
func byteOffsetContainingRune(r *runeIndex, b int) int {
	pos := 0
	for i, rn := range r.runes {
		width := len(string(rn))
		if b < pos+width {
			return i
		}
		pos += width
	}
	
	return r.count()
}

// runeCount returns the number of runes in a string.
func runeCount(s string) int {
	return len([]rune(s))
}
