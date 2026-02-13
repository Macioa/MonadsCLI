package types

// Node is a linked-list style node that encapsulates shape/block data from
// Lucid exports. Children maps route names (e.g. "Yes", "No") to child nodes.
// All fields are optional; nodes may have any set of children or none.
type Node struct {
	// Identity
	ID string `json:"id,omitempty"`

	// Label: shape/block type name (e.g. "Process", "Decision", "Predefined process")
	Label string `json:"label"`

	// Internal text: primary content
	Text string `json:"text"`

	// Text areas: label -> text for multiple regions (e.g. "Text" -> "Process")
	TextAreas map[string]string `json:"textAreas,omitempty"`

	// Tags (e.g. "NoValidation")
	Tags []string `json:"tags,omitempty"`

	// Status (e.g. "Draft")
	Status string `json:"status,omitempty"`

	// Metadata: custom variables, customData key-values (e.g. testprop, TestProp)
	Metadata map[string]string `json:"metadata,omitempty"`

	// Shape library (e.g. "Flowchart Shapes/Containers")
	ShapeLibrary string `json:"shapeLibrary,omitempty"`

	// Comments
	Comments string `json:"comments,omitempty"`

	// Children: route name -> child node (route names come from line labels)
	Children map[string]*Node `json:"children,omitempty"`
}
