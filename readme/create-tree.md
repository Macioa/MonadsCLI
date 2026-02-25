<div style="background-color: white;">

# Creating a Lucidchart Decision Tree

<p align="center"><strong>1. Open Lucidchart</strong> (<a href="https://lucid.app/documents">lucid.app/documents</a>) and create a new flowchart (empty doc).</p>

<p align="center"><img src="../images/blank_doc.png" alt="New blank document" /></p>

<p align="center"><strong>2. Create a terminator node</strong> and label it "Start".</p>

<p align="center"><img src="../images/terminator.png" alt="Terminator node" /></p>

<p align="center"><strong>3. Drag the node handles</strong> to add terminators, process nodes, decision nodes, and predefined processes, until the tree is complete.</p>

<p align="center"><img src="../images/node_handle.png" alt="Node handle" /></p>

<p align="center"><img src="../images/expand_tree.png" alt="Expand tree" /></p>

<p align="center"><strong>4. Export to CSV</strong> to save in project.</p>

<p align="center"><img src="../images/export.png" alt="Export to CSV" /></p>

---

# Node Types

<p align="center"><strong>Terminator</strong> – A terminator starts or stops the flow. Change the text to "Start" or "End".</p>

<p align="center"><img src="../images/terminator.png" alt="Terminator" /></p>

<p align="center"><strong>Process</strong> – A process node runs the text from the node with an AI CLI. After the process completes, a second CLI is used to automatically validate the changes made by the process. Unvalidated results will automatically be retried. See <a href="decision-tree-process.md">Decision tree process</a> for the full pipeline (run, validate, retry).</p>

<p align="center"><img src="../images/process.png" alt="Process node" /></p>

<p align="center"><strong>Decision</strong> – A multiple choice question posed to the LLM to determine which branch to follow. Any node with multiple children with labeled choice arrows will be treated as a decision node.</p>

<p align="center"><img src="../images/decision.png" alt="Decision node" /></p>

<p align="center"><strong>Predefined process</strong> – A predefined process runs the same as a regular process but can be used to distinguish prompts that use traditional deterministic tooling like scripts or MCPs.</p>

---

## Docs

- [Quick](quick.md)
- [Install](install.md)
- [Creating a Lucidchart decision tree](create-tree.md)
- [Metadata in trees](metadata.md)
- [Settings and keys](settings.md)

</div>
