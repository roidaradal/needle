package needle

var templateHTML = `
<!doctype html>
<html>
<head>
    <title>Needle | %ModuleName%</title>
    <style>
        table {
            border-top: 1px solid black;
            border-left: 1px solid black;
            border-collapse: collapse;
            margin: 1em;
        }
        th, td {
            min-width: 6em;
            padding: 5px;
            border-right: 1px solid black;
            border-bottom: 1px solid black;
        }
        td.center {
            text-align: center;
        }
        td.left {
            text-align: left;
        }
        td.right {
            text-align: right;
        }
        td.global {
            background-color: aqua;
            min-width: 1em; padding: 3px;
        }
        td.local {
            background-color: yellow;
            min-width: 1em; padding: 3px;
        }
        h1, h2 {
            padding: 0;
            margin: 5px;
            text-align: center;
        }
        .centered {
            text-align: center;
        }
        #header {
            width: 100%; height: 15%;
            border-bottom: 1px solid black;
        }
        #body {
            width: 100%; height: 85%;
        }
        #body table, #body div>ul{
            margin: 1em auto;
        }
        .hidden {
            display: none !important;
        }
        #mod, #code, #deps {
            width: 100%; height: 100%;
            overflow: auto;
        }
        #deps {
            text-align: center;
        }
        button.active {
            background-color: yellow;
            font-weight: bold;
        }
        #tabs, #tabs-mod, #tabs-code, #tabs-deps {
            width: 100%;
            display: flex;
            justify-content: center;
        }
        #tabs button, #tabs-deps button {
            width: 20%;
        }
        #tabs-mod button {
            width: 15%;
        }
        #tabs-code button {
            width: 10%;
        }
        button:hover {
            cursor: pointer;
        }
        canvas{
            border: 1px solid black;
        }
    </style>
</head>
<body><div id="app">
    <div id="header">
        <h1 class="centered">%ModuleName%</h1>
        <div id="tabs">
            <button id="btn-mod" onclick="changeTab('mod')" class="active">Module</button>
            <button id="btn-code" onclick="changeTab('code')">Code</button>
            <button id="btn-deps" onclick="changeTab('deps')">Dependencies</button>
        </div>
        <div id="tabs-mod">
            <button id="btn-mod-summary" onclick="changeSubTab('mod','summary')" class="active">Summary</button>
            <button id="btn-mod-files" onclick="changeSubTab('mod', 'files')">Files</button>
            <button id="btn-mod-lines" onclick="changeSubTab('mod', 'lines')">Lines</button>
            <button id="btn-mod-chars" onclick="changeSubTab('mod', 'chars')">Chars</button>
        </div>
        <div id="tabs-code" class="hidden">
            <button id="btn-code-summary" onclick="changeSubTab('code','summary')" class="active">Summary</button>
            <button id="btn-code-globals" onclick="changeSubTab('code', 'globals')">Globals</button>
            <button id="btn-code-functions" onclick="changeSubTab('code', 'functions')">Functions</button>
            <button id="btn-code-types" onclick="changeSubTab('code', 'types')">Types</button>
            <button id="btn-code-lines" onclick="changeSubTab('code', 'lines')">Lines</button>
            <button id="btn-code-chars" onclick="changeSubTab('code', 'chars')">Chars</button>
        </div>
        <div id="tabs-deps" class="hidden">
            <button id="btn-deps-dependent" onclick="changeSubTab('deps','dependent')" class="active">Dependent</button>
            <button id="btn-deps-independent" onclick="changeSubTab('deps', 'independent')">Independent</button>
            <button id="btn-deps-external" onclick="changeSubTab('deps', 'external')">External</button>
        </div>
    </div>

    <div id="body">     
        <div id="mod">
            <div id="mod-summary">
                <table><tbody>
                    <tr>
                        <th>Packages</th>
                        <th>%ModPackageCount%</th>
                        <td><b>Library</b><br/>%LibPackageCount%</td>
                        <td><b>Main</b><br/>%MainPackageCount%</td>
                    </tr>
                    <tr>
                        <th>Files</th>
                        <th>%ModFileCount%</th>
                        <td><b>Code</b><br/>%CodeFileCount%<br/>%CodeFileShare%</td>
                        <td><b>Test</b><br/>%TestFileCount%<br/>%TestFileShare%</td>
                    </tr>
                    <tr>
                        <th>Lines</th>
                        <th>%ModLineCount%</th>
                        <td><b>Code</b><br/>%CodeLineCount%<br/>%CodeLineShare%</td>
                        <td><b>Test</b><br/>%TestLineCount%<br/>%TestLineShare%</td>
                    </tr>
                    <tr>
                        <th>Characters</th>
                        <th>%ModCharCount%</th>
                        <td><b>Code</b><br/>%CodeCharCount%<br/>%CodeCharShare%</td>
                        <td><b>Test</b><br/>%TestCharCount%<br/>%TestCharShare%</td>
                    </tr>
                    <tr>
                        <th>AvgLinePerFile</th>
                        <th>%AvgLinePerFile%</th>
                        <td><b>Code</b><br/>%CodeALPF%</td>
                        <td><b>Test</b><br/>%TestALPF%</td>
                    </tr>
                    <tr>
                        <th>AvgCharPerFile</th>
                        <th>%AvgCharPerFile%</th>
                        <td><b>Code</b><br/>%CodeACPF%</td>
                        <td><b>Test</b><br/>%TestACPF%</td>
                    </tr>
                    <tr>
                        <th>AvgCharPerLine</th>
                        <th>%AvgCharPerLine%</th>
                        <td><b>Code</b><br/>%CodeACPL%</td>
                        <td><b>Test</b><br/>%TestACPL%</td>
                    </tr>
                </tbody></table>
            </div>

            <div id="mod-files" class="hidden">
                <table>
                    <thead><tr>
                        <th>Package</th>
                        <th>%</th>
                        <th>Files</th>
                        %ModuleTableHeader%
                        <th><button id="toggle-mod-files" onclick="toggleList('mod','files', 'Files')">Show Files</button></th>
                    </tr></thead>
                    <tbody>%ModuleTable%</tbody>
                </table>
            </div>

            <div id="mod-lines" class="hidden">
                <table>
                    <thead><tr>
                        <th>Package</th>
                        <th>%</th>
                        <th>Lines</th>
                        <th title="AvgLinePerFile">ALPF</th>
                        <th colspan="3"><button id="toggle-mod-lines" onclick="toggleList('mod', 'lines', 'File Lines')">Show File Lines</button></th>
                    </tr></thead>
                    <tbody>%LinesTable%</tbody>
                </table>
            </div>

            <div id="mod-chars" class="hidden">
                <table>
                    <thead><tr>
                        <th>Package</th>
                        <th>%</th>
                        <th>Chars</th>
                        <th title="AvgCharPerFile">ACPF</th>
                        <th title="AvgCharPerLine">ACPL</th>
                        <th colspan="3"><button id="toggle-mod-chars" onclick="toggleList('mod', 'chars', 'File Chars')">Show File Chars</button></th>
                    </tr></thead>
                    <tbody>%CharsTable%</tbody>
                </table>
            </div>
        </div>

        <div id="code" class="hidden">
            <div id="code-summary">
                <table><tbody>
                    <tr>
                        <th>Lines</th>
                        <td><b>Code</b><br/>%CodesLineCount%<br/>%CodesLineShare%</td>
                        <td><b>Error</b><br/>%ErrorLineCount%<br/>%ErrorLineShare%</td>
                        <td><b>Head</b><br/>%HeadLineCount%<br/>%HeadLineShare%</td>
                        <td><b>Comment</b><br/>%CommentLineCount%<br/>%CommentLineShare%</td>
                        <td><b>Space</b><br/>%SpaceLineCount%<br/>%SpaceLineShare%</td>
                        <td><b>Total</b><br/>%ModLineCount%<br/>&nbsp;</td>
                    </tr>
                    <tr>
                        <th>Characters</th>
                        <td><b>Code</b><br/>%CodesCharCount%<br/>%CodesCharShare%</td>
                        <td><b>Error</b><br/>%ErrorCharCount%<br/>%ErrorCharShare%</td>
                        <td><b>Head</b><br/>%HeadCharCount%<br/>%HeadCharShare%</td>
                        <td><b>Comment</b><br/>%CommentCharCount%<br/>%CommentCharShare%</td>
                        <td><b>Space</b><br/>%SpaceCharCount%<br/>%SpaceCharShare%</td>
                        <td><b>Total</b><br/>%ModCharCount%<br/>&nbsp;</td>
                    </tr>
                    <tr>
                        <th rowspan="2">Code</th>
                        <td colspan="2" class="center"><b>Globals:</b> %GlobalCount%</td>
                        <td colspan="2" class="center"><b>Functions:</b> %FunctionCount%</td>
                        <td colspan="2" class="center"><b>Types:</b> %TypeCount%</td>
                    </tr>
                    <tr>
                        <td><b>Public:</b> %PublicGlobalCount%</td>
                        <td><b>Private:</b> %PrivateGlobalCount%</td>
                        <td><b>Public:</b> %PublicFunctionCount%</td>
                        <td><b>Private:</b> %PrivateFunctionCount%</td>
                        <td><b>Public:</b> %PublicTypeCount%</td>
                        <td><b>Private:</b> %PrivateTypeCount%</td>
                    </tr>
                </tbody></table>
            </div>

            
            <div id="code-globals" class="hidden">
                <table>%GlobalsTable%</table>
            </div>

            <div id="code-functions" class="hidden">
                <table>%FunctionsTable%</table>
            </div>

            <div id="code-types" class="hidden">
                <table>%TypesTable%</table>
            </div>

            <div id="code-lines" class="hidden">
                <table>
                    <thead><tr>
                        <th>Packages</th>
                        
                        <th>Lines</th>
                        <th colspan="2">Codes</th>
                        <th colspan="2">Error</th>
                        <th colspan="2">Head</th>
                        <th colspan="2">Comment</th>
                        <th colspan="2">Space</th>
                        <th><button id="toggle-code-lines" onclick="toggleList('code', 'lines', '%')">Show %</button></th>
                    </tr></thead>
                    <tbody>%CodeLinesTable%</tbody>
                </table>
            </div>

            <div id="code-chars" class="hidden">
                <table>
                    <thead><tr>
                        <th>Packages</th>
                        <th>Chars</th>
                        <th colspan="2">Codes</th>
                        <th colspan="2">Error</th>
                        <th colspan="2">Head</th>
                        <th colspan="2">Comment</th>
                        <th colspan="2">Space</th>
                        <th><button id="toggle-code-chars" onclick="toggleList('code','chars','%')">Show %</button></th>
                    </tr></thead>
                    <tbody>%CodeCharsTable%</tbody>
                </table>
            </div>
        </div>
        
        <div id="deps" class="hidden">
            <div id="deps-dependent">
                <h2>Dependent: %DependentCount% / %ModPackageCount%</h2>
                %DependencyCanvas%
                <table id="deps-dependent-table">
                    %DependencyTable%
                </table>
            </div>

            <div id="deps-independent" class="hidden">
                <h2>Independent: %IndependentCount% / %ModPackageCount%</h2>
                <table><tbody>
                    %IndependentTable%
                </tbody></table>
            </div>

            <div id="deps-external" class="hidden">
                <h2>External: %ExternalDepsCount%</h2>
                <table>
                    %ExternalDepsTable%
                </table>
            </div>
        </div>
    </div>

    <script>
        const nodeRadius = 30;
        const nodes = {%DependencyNodes%};
        const edges = [%DependencyEdges%];
        var currentTab = 'mod';
        var currentSubTab = {
            'mod'   : 'summary',
            'code'  : 'summary',
            'deps'  : 'dependent',
        };
        var currentView = {
            'deps-dependent': 'table',
        };
        var isExpanded = {};
        var $id = function(id) { return document.getElementById(id) };
        var $class = function(className) { return Array.from(document.getElementsByClassName(className)) };
        function changeTab(tab) {
            if(tab == currentTab) {
                return;
            }
            $id('btn-' + currentTab).classList.remove('active');
            $id('btn-' + tab).classList.add('active');
            $id('tabs-' + currentTab).classList.add('hidden');
            $id('tabs-' + tab).classList.remove('hidden');
            $id(currentTab).classList.add('hidden');
            $id(tab).classList.remove('hidden');
            currentTab = tab;
        }
        function changeSubTab(tab, subTab) {
            let old = currentSubTab[tab];
            if(old == subTab) {
                return;
            }
            $id('btn-' + tab + '-' + old).classList.remove('active');
            $id('btn-' + tab + '-' + subTab).classList.add('active');
            $id(tab + '-' + old).classList.add('hidden');
            $id(tab + '-' + subTab).classList.remove('hidden');
            currentSubTab[tab] = subTab;
        }
        function toggleList(tab, subTab, name) {
            let key = tab + '-' + subTab;
            let expanded = isExpanded[key];
            if(expanded) {
                $id('toggle-'+key).innerHTML = 'Show ' + name;
                $class(key+'-list').forEach(function(elt){
                    elt.classList.add('hidden');
                });
            } else {
                $id('toggle-'+key).innerHTML = 'Hide ' + name;
                $class(key+'-list').forEach(function(elt){
                    elt.classList.remove('hidden');
                });
            }         
            isExpanded[key] = !expanded; // toggle
        }
        function drawNode(ctx, node, name) {
            ctx.beginPath();
            ctx.arc(node.x, node.y, nodeRadius, 0, Math.PI*2);
            ctx.fillStyle = node.sink ? '#FF0' : '#DDE';
            ctx.fill();
            ctx.strokeStyle = '#333';
            ctx.stroke();
            ctx.fillStyle = '#F00';
            ctx.textAlign = 'center';
            ctx.textBaseline = 'middle';
            ctx.fillText(name, node.x, node.y);
        }
        function drawEdge(ctx, node1, node2) {
            const angle = Math.atan2(node2.y-node1.y, node2.x-node1.x);
            const arrowSize = 10;
            const x1 = node1.x + Math.cos(angle) * nodeRadius;
            const y1 = node1.y + Math.sin(angle) * nodeRadius;
            const x2 = node2.x - Math.cos(angle) * nodeRadius;
            const y2 = node2.y - Math.sin(angle) * nodeRadius;

            ctx.beginPath();
            ctx.moveTo(x1,y1); ctx.lineTo(x2,y2);
            ctx.strokeStyle = '#555';
            ctx.stroke();

            ctx.save();
            ctx.translate(x2, y2); ctx.rotate(angle);
            ctx.beginPath();
            ctx.moveTo(0, 0);
            ctx.lineTo(-arrowSize, -arrowSize / 2);
            ctx.lineTo(-arrowSize, arrowSize / 2);
            ctx.closePath();
            ctx.fillStyle = '#555'; 
            ctx.fill();
            ctx.restore();
        }
        function drawGraph(ctx) {
            edges.forEach(edge => {
                console.log(edge);
                const parts = edge.split('-');
                const node1 = nodes[parts[0]];
                const node2 = nodes[parts[1]];
                if(node1 && node2) {
                    drawEdge(ctx, node1, node2);
                }
            });
            for(let key in nodes) {
                drawNode(ctx, nodes[key], key)
            }
        }
        function toggleDependentView() {
            if(currentView['deps-dependent'] == 'table') {
                $id('deps-dependent-graph').classList.remove('hidden');
                $id('deps-dependent-table').classList.add('hidden');
                $id('view-deps-dependent').innerHTML = 'Show Table';
                currentView['deps-dependent'] = 'graph';
            } else {
                $id('deps-dependent-graph').classList.add('hidden');
                $id('deps-dependent-table').classList.remove('hidden');
                $id('view-deps-dependent').innerHTML = 'Show Graph';
                currentView['deps-dependent'] = 'table';
            }
        }
        window.onload = function(){
            const canvas = $id('deps-dependent-graph');
            const ctx = canvas.getContext('2d');
            ctx.clearRect(0,0, canvas.width, canvas.height);
            drawGraph(ctx);
        };
    </script>
</div></body>
</html>
`
