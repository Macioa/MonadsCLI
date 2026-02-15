package document

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ryanmontgomery/MonadsCLI/types"
)

const unlabeledRoute = ""

// lucidJSON structures for unmarshaling Lucid API document contents
type lucidJSON struct {
	ID    string       `json:"id"`
	Title string       `json:"title"`
	Pages []lucidPage  `json:"pages"`
}

type lucidPage struct {
	Items lucidItems `json:"items"`
}

type lucidItems struct {
	Shapes []lucidShape `json:"shapes"`
	Lines  []lucidLine  `json:"lines"`
}

type lucidShape struct {
	ID         string         `json:"id"`
	Class      string         `json:"class"`
	TextAreas  []lucidTextArea `json:"textAreas"`
	CustomData []lucidKeyVal  `json:"customData"`
}

type lucidLine struct {
	Endpoint1  lucidEndpoint  `json:"endpoint1"`
	Endpoint2  lucidEndpoint  `json:"endpoint2"`
	TextAreas  []lucidTextArea `json:"textAreas"`
	CustomData []lucidKeyVal  `json:"customData"`
}

type lucidEndpoint struct {
	ConnectedTo string `json:"connectedTo"`
}

type lucidTextArea struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

type lucidKeyVal struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// classToLabel maps Lucid class names to display labels
func classToLabel(class string) string {
	switch class {
	case "ProcessBlock":
		return "Process"
	case "DecisionBlock":
		return "Decision"
	case "PredefinedProcessBlock":
		return "Predefined process"
	default:
		if idx := strings.Index(class, "Block"); idx > 0 {
			return class[:idx]
		}
		return class
	}
}

// TransformFromLucidJSON converts Lucid API document contents JSON into a Document with node tree.
func TransformFromLucidJSON(data []byte) (*types.Document, error) {
	var raw lucidJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal lucid JSON: %w", err)
	}

	doc := &types.Document{
		ID:    raw.ID,
		Title: raw.Title,
	}

	// Collect shapes and lines from first page
	var shapes []lucidShape
	var lines []lucidLine
	for _, p := range raw.Pages {
		shapes = append(shapes, p.Items.Shapes...)
		lines = append(lines, p.Items.Lines...)
	}

	// Build shape id -> node map
	nodes := make(map[string]*types.Node)
	for _, s := range shapes {
		n := shapeToNode(s)
		nodes[s.ID] = n
	}

	// Build adjacency: source id -> [(route, dest id)]
	outEdges := make(map[string][]struct{ route string; destID string })
	inDegree := make(map[string]int)
	for _, l := range lines {
		src := l.Endpoint1.ConnectedTo
		dst := l.Endpoint2.ConnectedTo
		route := unlabeledRoute
		for _, ta := range l.TextAreas {
			if t := strings.TrimSpace(ta.Text); t != "" {
				route = t
				break
			}
		}
		outEdges[src] = append(outEdges[src], struct{ route string; destID string }{route, dst})
		inDegree[dst]++
	}

	// Find root(s): nodes with in-degree 0
	var roots []string
	for id := range nodes {
		if inDegree[id] == 0 {
			roots = append(roots, id)
		}
	}
	if len(roots) == 0 {
		// No shapes or cyclic graph; return doc with nil root
		return doc, nil
	}

	// Build tree from first root via DFS
	doc.Root = buildTree(nodes, outEdges, roots[0], make(map[string]bool))
	return doc, nil
}

func shapeToNode(s lucidShape) *types.Node {
	textAreas := make(map[string]string)
	var primary string
	for _, ta := range s.TextAreas {
		textAreas[ta.Label] = ta.Text
		if ta.Label == "Text" || primary == "" {
			primary = ta.Text
		}
	}
	if primary == "" && len(textAreas) > 0 {
		for _, v := range textAreas {
			primary = v
			break
		}
	}
	metadata := make(map[string]string)
	for _, kv := range s.CustomData {
		metadata[kv.Key] = kv.Value
	}
	return &types.Node{
		ID:        s.ID,
		Label:     classToLabel(s.Class),
		Text:      primary,
		TextAreas: textAreas,
		Metadata:  metadata,
	}
}

func buildTree(nodes map[string]*types.Node, outEdges map[string][]struct{ route string; destID string }, id string, visited map[string]bool) *types.Node {
	if visited[id] {
		return nil
	}
	visited[id] = true
	n := nodes[id]
	if n == nil {
		return nil
	}
	edges := outEdges[id]
	if len(edges) == 0 {
		return n
	}
	n.Children = make(map[string]*types.Node)
	for _, e := range edges {
		child := buildTree(nodes, outEdges, e.destID, visited)
		if child != nil {
			route := e.route
			if route == "" {
				route = unlabeledRoute
			}
			// If multiple edges share same route, last wins (or we could merge - for tree we assume unique)
			n.Children[route] = child
		}
	}
	return n
}

// TransformFromCSV converts Lucid CSV export into a Document with node tree.
func TransformFromCSV(data []byte) (*types.Document, error) {
	r := csv.NewReader(strings.NewReader(string(data)))
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read CSV: %w", err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("CSV has no data rows")
	}

	header := rows[0]
	col := func(name string) int {
		for i, h := range header {
			if strings.TrimSpace(h) == name {
				return i
			}
		}
		return -1
	}
	idxID := col("Id")
	idxName := col("Name")
	idxShapeLib := col("Shape Library")
	idxLineSrc := col("Line Source")
	idxLineDst := col("Line Destination")
	idxTags := col("Tags")
	idxStatus := col("Status")
	idxText1 := col("Text Area 1")
	idxComments := col("Comments")
	idxTestprop := col("testprop")

	if idxID < 0 || idxName < 0 {
		return nil, fmt.Errorf("CSV missing required columns (Id, Name)")
	}

	doc := &types.Document{}
	nodes := make(map[string]*types.Node)
	var lines []struct{ src, dst, route string }
	inDegree := make(map[string]int)

	for _, row := range rows[1:] {
		if len(row) <= idxID {
			continue
		}
		id := strings.TrimSpace(row[idxID])
		name := strings.TrimSpace(safeAt(row, idxName))
		if id == "" || name == "" {
			continue
		}

		// Document row
		if name == "Document" {
			doc.Title = strings.TrimSpace(safeAt(row, idxText1))
			if s := strings.TrimSpace(safeAt(row, idxStatus)); s != "" {
				doc.Status = s
			}
			continue
		}

		// Page row - skip
		if name == "Page" {
			continue
		}

		// Line row
		if name == "Line" {
			src := strings.TrimSpace(safeAt(row, idxLineSrc))
			dst := strings.TrimSpace(safeAt(row, idxLineDst))
			route := unlabeledRoute
			if idxTags >= 0 && idxTags < len(row) {
				tags := strings.TrimSpace(row[idxTags])
				if tags != "" {
					route = tags
				}
			}
			if idxText1 >= 0 && idxText1 < len(row) && route == unlabeledRoute {
				t1 := strings.TrimSpace(row[idxText1])
				if t1 != "" {
					route = t1
				}
			}
			if src != "" && dst != "" {
				lines = append(lines, struct{ src, dst, route string }{src, dst, route})
				inDegree[dst]++
			}
			continue
		}

		// Shape row
		n := &types.Node{
			ID:    id,
			Label: name,
			Text:  strings.TrimSpace(safeAt(row, idxText1)),
		}
		if idxShapeLib >= 0 {
			n.ShapeLibrary = strings.TrimSpace(safeAt(row, idxShapeLib))
		}
		if idxComments >= 0 {
			n.Comments = strings.TrimSpace(safeAt(row, idxComments))
		}
		if idxStatus >= 0 {
			n.Status = strings.TrimSpace(safeAt(row, idxStatus))
		}
		if idxTags >= 0 {
			tags := strings.TrimSpace(safeAt(row, idxTags))
			if tags != "" {
				n.Tags = []string{tags}
			}
		}
		if idxTestprop >= 0 {
			tp := strings.TrimSpace(safeAt(row, idxTestprop))
			if tp != "" {
				n.Metadata = map[string]string{"testprop": tp}
			}
		}
		nodes[id] = n
	}

	// Build outEdges from lines (use row Id for shapes, Line Source/Dest reference shape Id from CSV)
	// In CSV, Line Source and Line Destination are the Id values (3,4,5,6,7)
	outEdges := make(map[string][]struct{ route string; destID string })
	for _, l := range lines {
		outEdges[l.src] = append(outEdges[l.src], struct{ route string; destID string }{l.route, l.dst})
	}

	// Find roots
	var roots []string
	for id := range nodes {
		if inDegree[id] == 0 {
			roots = append(roots, id)
		}
	}
	if len(roots) == 0 {
		return doc, nil
	}

	visited := make(map[string]bool)
	doc.Root = buildTree(nodes, outEdges, roots[0], visited)
	return doc, nil
}

func safeAt(row []string, i int) string {
	if i < 0 || i >= len(row) {
		return ""
	}
	return row[i]
}

// csvHeader matches the Lucid CSV export format.
const csvHeader = "Id,Name,Shape Library,Page ID,Contained By,Group,Line Source,Line Destination,Source Arrow,Destination Arrow,Tags,Status,Text Area 1,Comments,testprop"

// edge represents a parent->child link with route label.
type edge struct {
	srcID  int
	dstID  int
	route  string
}

// TransformToCSV converts a Document back to Lucid CSV export format.
func TransformToCSV(doc *types.Document) ([]byte, error) {
	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}

	var buf strings.Builder
	w := csv.NewWriter(&buf)
	if err := w.Write(strings.Split(csvHeader, ",")); err != nil {
		return nil, err
	}

	nextID := 1
	nodeToID := make(map[*types.Node]int)
	var edges []edge

	// Assign IDs and collect edges via DFS
	var collect func(n *types.Node)
	collect = func(n *types.Node) {
		if n == nil {
			return
		}
		nodeToID[n] = nextID
		nextID++
		for route, child := range n.Children {
			if child != nil {
				collect(child)
				edges = append(edges, edge{nodeToID[n], nodeToID[child], route})
			}
		}
	}
	collect(doc.Root)

	// Document row (Id=1)
	status := doc.Status
	if status == "" {
		status = "Draft"
	}
	if err := w.Write([]string{"1", "Document", "", "", "", "", "", "", "", "", "", status, doc.Title, "", ""}); err != nil {
		return nil, err
	}
	// Page row (Id=2)
	if err := w.Write([]string{"2", "Page", "", "", "", "", "", "", "", "", "", "", "Page 1", "", ""}); err != nil {
		return nil, err
	}

	// Shape rows
	var writeNode func(n *types.Node)
	writeNode = func(n *types.Node) {
		if n == nil {
			return
		}
		id := nodeToID[n]
		tags := ""
		if len(n.Tags) > 0 {
			tags = n.Tags[0]
		}
		testprop := ""
		if n.Metadata != nil {
			testprop = n.Metadata["testprop"]
		}
		row := []string{
			fmt.Sprint(id), n.Label, n.ShapeLibrary, "2", "", "",
			"", "", "", "", tags, n.Status, n.Text, n.Comments, testprop,
		}
		if err := w.Write(row); err != nil {
			panic(err)
		}
		for _, child := range n.Children {
			writeNode(child)
		}
	}
	writeNode(doc.Root)

	// Line rows
	for _, e := range edges {
		tags := e.route
		row := []string{
			fmt.Sprint(nextID), "Line", "", "2", "", "",
			fmt.Sprint(e.srcID), fmt.Sprint(e.dstID), "None", "Arrow",
			tags, "", "", "", "",
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
		nextID++
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return []byte(strings.TrimSpace(buf.String())), nil
}
