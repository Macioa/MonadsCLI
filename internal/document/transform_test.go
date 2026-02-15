package document

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ryanmontgomery/MonadsCLI/types"
)

func testdata(path string) []byte {
	b, err := os.ReadFile(filepath.Join("..", "..", "test", path))
	if err != nil {
		panic(err)
	}
	return b
}

func TestTransformFromLucidJSON(t *testing.T) {
	data := testdata("sample_data.json")
	doc, err := TransformFromLucidJSON(data)
	if err != nil {
		t.Fatalf("TransformFromLucidJSON: %v", err)
	}
	if doc.ID != "35fbd25a-3992-41c0-9d47-e1f9c86ea116" {
		t.Errorf("doc.ID = %q, want 35fbd25a-3992-41c0-9d47-e1f9c86ea116", doc.ID)
	}
	if doc.Title != "Blank diagram" {
		t.Errorf("doc.Title = %q, want Blank diagram", doc.Title)
	}
	if doc.Root == nil {
		t.Fatal("doc.Root is nil")
	}
	// Root: Process (CXSt0TzudBZw)
	if doc.Root.Label != "Process" {
		t.Errorf("root.Label = %q, want Process", doc.Root.Label)
	}
	if doc.Root.Text != "Process" {
		t.Errorf("root.Text = %q, want Process", doc.Root.Text)
	}
	if len(doc.Root.Metadata) == 0 || doc.Root.Metadata["TestProp"] != "Test" {
		t.Errorf("root.Metadata = %v, want TestProp=Test", doc.Root.Metadata)
	}
	// Root has one unlabeled child: Predefined process
	if doc.Root.Children == nil {
		t.Fatal("root.Children is nil")
	}
	child, ok := doc.Root.Children[""]
	if !ok || child == nil {
		t.Fatalf("root.Children[unlabeled] missing or nil, have keys: %v", keys(doc.Root.Children))
	}
	if child.Label != "Predefined process" {
		t.Errorf("child.Label = %q, want Predefined process", child.Label)
	}
	// That child has one unlabeled child: Decision
	decision := child.Children[""]
	if decision == nil || decision.Label != "Decision" {
		t.Errorf("decision node: got %v", decision)
	}
	// Decision has Yes -> Predefined process, No -> Process
	if decision.Children["Yes"] == nil || decision.Children["Yes"].Label != "Predefined process" {
		t.Errorf("decision.Children[Yes] = %v", decision.Children["Yes"])
	}
	if decision.Children["No"] == nil || decision.Children["No"].Label != "Process" {
		t.Errorf("decision.Children[No] = %v", decision.Children["No"])
	}
}

func TestTransformFromCSV(t *testing.T) {
	data := testdata("sample_data.csv")
	doc, err := TransformFromCSV(data)
	if err != nil {
		t.Fatalf("TransformFromCSV: %v", err)
	}
	if doc.Title != "Blank diagram" {
		t.Errorf("doc.Title = %q, want Blank diagram", doc.Title)
	}
	if doc.Status != "Draft" {
		t.Errorf("doc.Status = %q, want Draft", doc.Status)
	}
	if doc.Root == nil {
		t.Fatal("doc.Root is nil")
	}
	// Root: Process (id 3)
	if doc.Root.Label != "Process" {
		t.Errorf("root.Label = %q, want Process", doc.Root.Label)
	}
	if doc.Root.Text != "Process" {
		t.Errorf("root.Text = %q, want Process", doc.Root.Text)
	}
	if !reflect.DeepEqual(doc.Root.Tags, []string{"NoValidation"}) {
		t.Errorf("root.Tags = %v, want [NoValidation]", doc.Root.Tags)
	}
	if doc.Root.Metadata["testprop"] != "Test" {
		t.Errorf("root.Metadata = %v, want testprop=Test", doc.Root.Metadata)
	}
	// Root -> Predefined process (unlabeled)
	child, ok := doc.Root.Children[""]
	if !ok || child == nil {
		t.Fatalf("root.Children[unlabeled] missing, have keys: %v", keys(doc.Root.Children))
	}
	if child.Label != "Predefined process" {
		t.Errorf("child.Label = %q, want Predefined process", child.Label)
	}
	// Predefined process -> Decision
	decision := child.Children[""]
	if decision == nil || decision.Label != "Decision" {
		t.Errorf("decision node: got %v", decision)
	}
	// Decision: No -> Process, Yes -> Predefined process
	if decision.Children["No"] == nil || decision.Children["No"].Label != "Process" {
		t.Errorf("decision.Children[No] = %v", decision.Children["No"])
	}
	if decision.Children["Yes"] == nil || decision.Children["Yes"].Label != "Predefined process" {
		t.Errorf("decision.Children[Yes] = %v", decision.Children["Yes"])
	}
}

func TestTransformFromLucidJSON_Empty(t *testing.T) {
	doc, err := TransformFromLucidJSON([]byte(`{"id":"x","title":"Empty","pages":[{"items":{"shapes":[],"lines":[]}}]}`))
	if err != nil {
		t.Fatalf("empty shapes should not error: %v", err)
	}
	if doc.Root != nil {
		t.Error("document with no shapes should have nil root")
	}
}

func TestTransformFromLucidJSON_Invalid(t *testing.T) {
	_, err := TransformFromLucidJSON([]byte(`{`))
	if err == nil {
		t.Fatal("invalid JSON should error")
	}
}

func TestTransformFromCSV_Empty(t *testing.T) {
	_, err := TransformFromCSV([]byte("Id,Name\n"))
	if err == nil {
		t.Fatal("CSV with no data rows should error")
	}
}

func TestTransformFromCSV_Invalid(t *testing.T) {
	_, err := TransformFromCSV([]byte(`Id,Name
3,Process,"unclosed`))
	if err == nil {
		t.Fatal("invalid CSV should error")
	}
}

func TestTransformToCSV_NilDocument(t *testing.T) {
	_, err := TransformToCSV(nil)
	if err == nil {
		t.Fatal("TransformToCSV(nil) should error")
	}
}

func TestTransformToCSV(t *testing.T) {
	// Roundtrip: CSV -> Document -> CSV -> Document
	data := testdata("sample_data.csv")
	doc1, err := TransformFromCSV(data)
	if err != nil {
		t.Fatalf("TransformFromCSV: %v", err)
	}
	csvOut, err := TransformToCSV(doc1)
	if err != nil {
		t.Fatalf("TransformToCSV: %v", err)
	}
	doc2, err := TransformFromCSV(csvOut)
	if err != nil {
		t.Fatalf("TransformFromCSV (roundtrip): %v", err)
	}
	// Verify structure preserved
	if doc1.Title != doc2.Title {
		t.Errorf("title: %q != %q", doc1.Title, doc2.Title)
	}
	if !nodeEqual(doc1.Root, doc2.Root) {
		t.Error("root node structure differs after roundtrip")
	}
}

// TestTransformRoundtripRetention verifies that the full cycle (CSV -> Document -> CSV)
// retains parent-child relationships, arrow/route labels, tags, and variables.
func TestTransformRoundtripRetention(t *testing.T) {
	data := testdata("sample_data.csv")
	doc, err := TransformFromCSV(data)
	if err != nil {
		t.Fatalf("TransformFromCSV: %v", err)
	}
	csvOut, err := TransformToCSV(doc)
	if err != nil {
		t.Fatalf("TransformToCSV: %v", err)
	}

	rows, err := csv.NewReader(strings.NewReader(string(csvOut))).ReadAll()
	if err != nil {
		t.Fatalf("parse output CSV: %v", err)
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
	idxName := col("Name")
	idxTags := col("Tags")
	idxLineSrc := col("Line Source")
	idxLineDst := col("Line Destination")
	idxTestprop := col("testprop")

	// 1. Verify tags and variables on root shape (Process with NoValidation, testprop=Test)
	var foundRootWithTags bool
	for _, row := range rows[1:] {
		if len(row) <= idxName {
			continue
		}
		name := strings.TrimSpace(row[idxName])
		if name == "Document" || name == "Page" || name == "Line" {
			continue
		}
		tags := ""
		if idxTags >= 0 && idxTags < len(row) {
			tags = strings.TrimSpace(row[idxTags])
		}
		testprop := ""
		if idxTestprop >= 0 && idxTestprop < len(row) {
			testprop = strings.TrimSpace(row[idxTestprop])
		}
		if name == "Process" && tags == "NoValidation" && testprop == "Test" {
			foundRootWithTags = true
			break
		}
	}
	if !foundRootWithTags {
		t.Error("roundtrip CSV: root Process should retain Tags=NoValidation and testprop=Test")
	}

	// 2. Verify parent-child relationships and arrow text (route labels) on lines
	// Expect: 4 lines total; 2 unlabeled, 1 with Tags=Yes, 1 with Tags=No
	var lineRows [][]string
	for _, row := range rows[1:] {
		if len(row) <= idxName {
			continue
		}
		if strings.TrimSpace(row[idxName]) == "Line" {
			lineRows = append(lineRows, row)
		}
	}
	if len(lineRows) != 4 {
		t.Errorf("roundtrip CSV: expected 4 line rows, got %d", len(lineRows))
	}
	routeCounts := make(map[string]int)
	for _, row := range lineRows {
		src := ""
		dst := ""
		route := ""
		if idxLineSrc >= 0 && idxLineSrc < len(row) {
			src = strings.TrimSpace(row[idxLineSrc])
		}
		if idxLineDst >= 0 && idxLineDst < len(row) {
			dst = strings.TrimSpace(row[idxLineDst])
		}
		if idxTags >= 0 && idxTags < len(row) {
			route = strings.TrimSpace(row[idxTags])
		}
		if src == "" || dst == "" {
			t.Errorf("roundtrip CSV: line missing Line Source or Line Destination: %v", row)
		}
		if route == "" {
			route = "(unlabeled)"
		}
		routeCounts[route]++
	}
	if routeCounts["Yes"] != 1 {
		t.Errorf("roundtrip CSV: expected 1 line with Tags=Yes (arrow text), got %d", routeCounts["Yes"])
	}
	if routeCounts["No"] != 1 {
		t.Errorf("roundtrip CSV: expected 1 line with Tags=No (arrow text), got %d", routeCounts["No"])
	}
	if routeCounts["(unlabeled)"] != 2 {
		t.Errorf("roundtrip CSV: expected 2 unlabeled lines, got %d", routeCounts["(unlabeled)"])
	}

	// 3. Verify Document structure (root has children with correct route keys)
	assertNodeRelationships(t, doc)
}

// TestTransformNodeRelationshipsTripleRoundtrip runs CSV->Doc->CSV three times and
// verifies node relationships (parent-child, route keys) are preserved at each step.
func TestTransformNodeRelationshipsTripleRoundtrip(t *testing.T) {
	data := testdata("sample_data.csv")
	doc, err := TransformFromCSV(data)
	if err != nil {
		t.Fatalf("TransformFromCSV: %v", err)
	}
	// Three full roundtrips
	for i := 0; i < 3; i++ {
		csvOut, err := TransformToCSV(doc)
		if err != nil {
			t.Fatalf("TransformToCSV roundtrip %d: %v", i+1, err)
		}
		doc, err = TransformFromCSV(csvOut)
		if err != nil {
			t.Fatalf("TransformFromCSV roundtrip %d: %v", i+1, err)
		}
		assertNodeRelationships(t, doc)
	}
}

// TestTransformCSVLineConnectivity verifies that the output CSV lines form the
// correct graph: root->Predefined->Decision->(Process via No, Predefined via Yes).
func TestTransformCSVLineConnectivity(t *testing.T) {
	data := testdata("sample_data.csv")
	doc, err := TransformFromCSV(data)
	if err != nil {
		t.Fatalf("TransformFromCSV: %v", err)
	}
	csvOut, err := TransformToCSV(doc)
	if err != nil {
		t.Fatalf("TransformToCSV: %v", err)
	}
	// Parse and build graph from CSV
	rows, _ := csv.NewReader(strings.NewReader(string(csvOut))).ReadAll()
	header := rows[0]
	col := func(name string) int {
		for i, h := range header {
			if strings.TrimSpace(h) == name {
				return i
			}
		}
		return -1
	}
	idxID, idxName := col("Id"), col("Name")
	idxLineSrc := col("Line Source")
	idxLineDst := col("Line Destination")
	idxTags := col("Tags")

	shapeByID := make(map[string]string) // id -> Label
	var inDegree map[string]int = make(map[string]int)
	outEdges := make(map[string][]struct{ dst string; route string })

	for _, row := range rows[1:] {
		if len(row) <= idxName {
			continue
		}
		id := strings.TrimSpace(row[idxID])
		name := strings.TrimSpace(row[idxName])
		if name == "Document" || name == "Page" {
			continue
		}
		if name == "Line" {
			src := strings.TrimSpace(safeAt(row, idxLineSrc))
			dst := strings.TrimSpace(safeAt(row, idxLineDst))
			route := strings.TrimSpace(safeAt(row, idxTags))
			if src != "" && dst != "" {
				outEdges[src] = append(outEdges[src], struct{ dst string; route string }{dst, route})
				inDegree[dst]++
			}
			continue
		}
		shapeByID[id] = name
	}
	// Find root (in-degree 0)
	var rootID string
	for id := range shapeByID {
		if inDegree[id] == 0 {
			if rootID != "" {
				t.Fatalf("multiple roots: %s and %s", rootID, id)
			}
			rootID = id
		}
	}
	if rootID == "" {
		t.Fatal("no root found in output CSV")
	}
	if shapeByID[rootID] != "Process" {
		t.Errorf("root shape: got %q, want Process", shapeByID[rootID])
	}
	// Root -> Predefined process (unlabeled)
	rootEdges := outEdges[rootID]
	if len(rootEdges) != 1 {
		t.Fatalf("root should have 1 outgoing edge, got %d", len(rootEdges))
	}
	if rootEdges[0].route != "" {
		t.Errorf("root edge should be unlabeled, got route %q", rootEdges[0].route)
	}
	midID := rootEdges[0].dst
	if shapeByID[midID] != "Predefined process" {
		t.Errorf("root child: got %q, want Predefined process", shapeByID[midID])
	}
	// Predefined process -> Decision (unlabeled)
	midEdges := outEdges[midID]
	if len(midEdges) != 1 {
		t.Fatalf("Predefined process should have 1 outgoing edge, got %d", len(midEdges))
	}
	decisionID := midEdges[0].dst
	if shapeByID[decisionID] != "Decision" {
		t.Errorf("decision shape: got %q", shapeByID[decisionID])
	}
	// Decision -> Process (No) and Predefined process (Yes)
	decEdges := outEdges[decisionID]
	if len(decEdges) != 2 {
		t.Fatalf("Decision should have 2 outgoing edges, got %d", len(decEdges))
	}
	routes := make(map[string]string)
	for _, e := range decEdges {
		routes[e.route] = shapeByID[e.dst]
	}
	if routes["No"] != "Process" {
		t.Errorf("Decision No edge: got %q, want Process", routes["No"])
	}
	if routes["Yes"] != "Predefined process" {
		t.Errorf("Decision Yes edge: got %q, want Predefined process", routes["Yes"])
	}
}

// TestTransformJSONToCSVRoundtrip verifies JSON->Doc->CSV->Doc preserves node relationships.
func TestTransformJSONToCSVRoundtrip(t *testing.T) {
	jsonData := testdata("sample_data.json")
	doc, err := TransformFromLucidJSON(jsonData)
	if err != nil {
		t.Fatalf("TransformFromLucidJSON: %v", err)
	}
	csvOut, err := TransformToCSV(doc)
	if err != nil {
		t.Fatalf("TransformToCSV: %v", err)
	}
	doc2, err := TransformFromCSV(csvOut)
	if err != nil {
		t.Fatalf("TransformFromCSV: %v", err)
	}
	assertNodeRelationships(t, doc2)
}

func assertNodeRelationships(t *testing.T, doc *types.Document) {
	t.Helper()
	if doc.Root == nil {
		t.Fatal("doc.Root is nil")
	}
	if doc.Root.Label != "Process" {
		t.Errorf("root.Label = %q, want Process", doc.Root.Label)
	}
	if doc.Root.Children[""] == nil || doc.Root.Children[""].Label != "Predefined process" {
		t.Error("root should have unlabeled child Predefined process")
	}
	mid := doc.Root.Children[""]
	decision := mid.Children[""]
	if decision == nil || decision.Label != "Decision" {
		t.Error("Predefined process should have unlabeled child Decision")
	}
	if decision.Children["Yes"] == nil || decision.Children["Yes"].Label != "Predefined process" {
		t.Error("Decision should have Yes child Predefined process")
	}
	if decision.Children["No"] == nil || decision.Children["No"].Label != "Process" {
		t.Error("Decision should have No child Process")
	}
}

func nodeEqual(a, b *types.Node) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Label != b.Label || a.Text != b.Text {
		return false
	}
	if !reflect.DeepEqual(a.Tags, b.Tags) {
		return false
	}
	if !reflect.DeepEqual(a.Metadata, b.Metadata) {
		return false
	}
	if len(a.Children) != len(b.Children) {
		return false
	}
	for k := range a.Children {
		if b.Children[k] == nil {
			return false
		}
		if !nodeEqual(a.Children[k], b.Children[k]) {
			return false
		}
	}
	return true
}

func keys(m map[string]*types.Node) []string {
	var k []string
	for x := range m {
		k = append(k, x)
	}
	return k
}
