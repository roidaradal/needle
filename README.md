# Needle
Go module analysis tool that shows dependency tree, module statistics, composition, and public API of packages.

The `needle` name comes from the tool's original intention: building a directed acyclic graph (DAG) of internal package dependencies.

DAG of Module ⇒ DAGoM ⇒ Needle

## Installation 
`go install github.com/roidaradal/needle@latest`

## Usage 
`needle <modulePath> (<reportPath>)`