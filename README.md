# Needle
Go module analysis tool that shows dependency tree, module statistics, composition, and public API of packages.

The `needle` name comes from the tool's original intention: building a directed acyclic graph (DAG) of internal package dependencies.

DAG of Module ⇒ DAGoM ⇒ Needle

## Installation 
`go install github.com/roidaradal/needle@latest`

OR 

Download `needle.exe` from the [releases](https://github.com/roidaradal/needle/releases) page. Add the folder where you saved `pson.exe` to your system PATH.

## Usage 
`needle <modulePath>`