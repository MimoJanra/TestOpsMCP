package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// ext-apps bundle fetching
// ---------------------------------------------------------------------------

const (
	launchDashboardURI       = "ui://widgets/launch-dashboard"
	launchDashboardURIPrefix = launchDashboardURI + "?launch_id="
	launchDashboardName      = "Launch Dashboard"
	widgetMimeType           = "text/html;profile=mcp-app"
)

func launchDashboardURIFor(launchID int64) string {
	return fmt.Sprintf("%s?launch_id=%d", launchDashboardURI, launchID)
}

func parseLaunchDashboardURI(uri string) (int64, bool) {
	if !strings.HasPrefix(uri, launchDashboardURIPrefix) {
		return 0, false
	}
	var id int64
	if _, err := fmt.Sscanf(uri[len(launchDashboardURIPrefix):], "%d", &id); err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// extAppsCandidates lists CDN URLs for the ext-apps browser bundle, tried in order.
var extAppsCandidates = []string{
	"https://unpkg.com/@modelcontextprotocol/ext-apps/dist/src/app-with-deps.js",
	"https://cdn.jsdelivr.net/npm/@modelcontextprotocol/ext-apps/dist/src/app-with-deps.js",
	"https://unpkg.com/@modelcontextprotocol/ext-apps/dist/app-with-deps.js",
	"https://unpkg.com/@modelcontextprotocol/ext-apps/app-with-deps.js",
	"https://cdn.jsdelivr.net/npm/@modelcontextprotocol/ext-apps/dist/app-with-deps.js",
}

// bundleFallback is a minimal stub used when the real ext-apps bundle cannot be
// fetched. It implements enough of the App API for the widget to function.
const bundleFallback = `globalThis.ExtApps={App:class{constructor(i,c,o){this._opts=o||{}}` +
	`set ontoolresult(f){this._tr=f}` +
	`set onhostcontextchanged(f){this._hcc=f}` +
	`set ontoolinput(f){this._ti=f}` +
	`async connect(){` +
	`window.addEventListener('message',e=>{` +
	`const d=e.data;` +
	`if(d&&d.type==='toolresult')this._tr?.({content:d.content||[]});` +
	`if(d&&d.type==='hostcontext')this._hcc?.(d.context);` +
	`});` +
	`window.parent?.postMessage({type:'mcp:widget:ready'},'*');` +
	`}` +
	`sendMessage(m){window.parent?.postMessage({type:'mcp:widget:message',payload:m},'*')}` +
	`getHostContext(){return null}` +
	`updateModelContext(){}` +
	`async callServerTool(){return{content:[]}}` +
	`openLink(o){try{window.open(o.url,'_blank')}catch(_){}}` +
	`downloadFile(){}` +
	`requestDisplayMode(){}` +
	`}};`

var (
	bundleOnce sync.Once
	bundleJS   string
)

// getExtAppsBundle returns the ext-apps browser bundle with ESM exports rewritten
// to globalThis.ExtApps. It is fetched once and cached for the process lifetime.
func getExtAppsBundle(logger interface {
	Info(string, any)
	Warn(string, any)
}) string {
	bundleOnce.Do(func() {
		client := &http.Client{Timeout: 15 * time.Second}
		for _, u := range extAppsCandidates {
			resp, err := client.Get(u)
			if err != nil {
				continue
			}
			body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<20)) // 4 MiB cap
			resp.Body.Close()
			if readErr != nil || resp.StatusCode != http.StatusOK || len(body) == 0 {
				continue
			}
			rewritten := rewriteESMExports(string(body))
			if rewritten != "" {
				bundleJS = rewritten
				if logger != nil {
					logger.Info("ext-apps bundle loaded", map[string]any{
						"url":  u,
						"size": len(bundleJS),
					})
				}
				return
			}
		}
		bundleJS = bundleFallback
		if logger != nil {
			logger.Warn("ext-apps bundle unavailable, using fallback stub", nil)
		}
	})
	return bundleJS
}

// rewriteESMExports converts trailing ESM export statements to globalThis.ExtApps = {...}.
// e.g. `export{Foo, Bar as Baz}` → `globalThis.ExtApps={Foo:Foo, Baz:Bar};`
func rewriteESMExports(js string) string {
	re := regexp.MustCompile(`export\s*\{([^}]+)\}\s*;?\s*$`)
	result := re.ReplaceAllStringFunc(js, func(match string) string {
		inner := re.FindStringSubmatch(match)[1]
		parts := strings.Split(inner, ",")
		props := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			halves := strings.SplitN(p, " as ", 2)
			if len(halves) == 2 {
				local := strings.TrimSpace(halves[0])
				exported := strings.TrimSpace(halves[1])
				props = append(props, exported+":"+local)
			} else {
				props = append(props, p+":"+p)
			}
		}
		return "globalThis.ExtApps={" + strings.Join(props, ",") + "};"
	})
	// Return empty string if no export was rewritten AND the bundle doesn't
	// already set globalThis.ExtApps (e.g. a CJS/UMD build).
	if result == js && !strings.Contains(js, "globalThis.ExtApps") {
		return ""
	}
	return result
}

// ---------------------------------------------------------------------------
// Launch Dashboard HTML template
// ---------------------------------------------------------------------------

const launchDashboardTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<style>
:root{--bg:#fff;--card:#f8f9fa;--text:#1a1a1a;--sub:#6b7280;--border:#e5e7eb;
  --passed:#16a34a;--passed-bg:#dcfce7;
  --failed:#dc2626;--failed-bg:#fee2e2;
  --broken:#d97706;--broken-bg:#fef3c7;
  --skip:#6b7280;--skip-bg:#f3f4f6;
  --run:#2563eb;--run-bg:#dbeafe}
html.dark{--bg:#1c1c1e;--card:#2c2c2e;--text:#f0f0f0;--sub:#9ca3af;--border:#3a3a3c;
  --passed-bg:#14532d;--failed-bg:#7f1d1d;--broken-bg:#78350f;--skip-bg:#374151;--run-bg:#1e3a8a}
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;
  background:var(--bg);color:var(--text);padding:12px;font-size:14px;line-height:1.4}
.card{background:var(--card);border:1px solid var(--border);border-radius:12px;overflow:hidden}
.header{padding:14px 16px;border-bottom:1px solid var(--border);display:flex;align-items:flex-start;gap:10px}
.launch-name{font-size:15px;font-weight:600;flex:1;word-break:break-word}
.badge{padding:3px 10px;border-radius:999px;font-size:11px;font-weight:700;letter-spacing:.04em;white-space:nowrap;flex-shrink:0}
.RUNNING{background:var(--run-bg);color:var(--run)}
.DONE{background:var(--passed-bg);color:var(--passed)}
.FAILED{background:var(--failed-bg);color:var(--failed)}
.CLOSED{background:var(--skip-bg);color:var(--skip)}
.UNKNOWN{background:var(--skip-bg);color:var(--skip)}
.prog{padding:14px 16px;border-bottom:1px solid var(--border)}
.prog-row{display:flex;justify-content:space-between;margin-bottom:8px;font-size:12px;color:var(--sub)}
.track{height:8px;background:var(--border);border-radius:99px;overflow:hidden}
.fill{height:100%;border-radius:99px;background:var(--passed);transition:width .4s ease}
.fill.running{background:var(--run)}
.stats{display:grid;grid-template-columns:repeat(4,1fr);border-bottom:1px solid var(--border)}
.stat{padding:14px 8px;text-align:center;border-right:1px solid var(--border)}
.stat:last-child{border-right:none}
.sv{font-size:20px;font-weight:700}
.sl{font-size:10px;color:var(--sub);margin-top:2px;text-transform:uppercase;letter-spacing:.06em}
.p{color:var(--passed)}.f{color:var(--failed)}.b{color:var(--broken)}.s{color:var(--skip)}
.meta{padding:10px 16px;font-size:12px;color:var(--sub);display:flex;gap:12px;flex-wrap:wrap;border-bottom:1px solid var(--border)}
.actions{padding:12px 16px;display:flex;gap:8px;flex-wrap:wrap}
.btn{padding:7px 13px;border-radius:8px;border:1px solid var(--border);background:var(--bg);
  color:var(--text);cursor:pointer;font-size:13px;font-weight:500;transition:.15s}
.btn:hover{background:var(--border)}.btn:active{opacity:.8}
.primary{background:#2563eb;color:#fff;border-color:#2563eb}
.primary:hover{background:#1d4ed8;border-color:#1d4ed8}
.danger{background:var(--failed-bg);color:var(--failed);border-color:var(--failed)}
.danger:hover{opacity:.8}
.loading,.error{padding:32px;text-align:center;color:var(--sub)}
.error{color:var(--failed)}
.spin{display:inline-block;width:20px;height:20px;
  border:2px solid var(--border);border-top-color:var(--run);
  border-radius:50%;animation:spin .8s linear infinite;margin-bottom:8px}
@keyframes spin{to{transform:rotate(360deg)}}
</style>
</head>
<body>
<div id="root"><div class="loading"><div class="spin"></div><div>Loading dashboard…</div></div></div>
<script>
/*__EXT_APPS_BUNDLE__*/
const{App}=globalThis.ExtApps;
const root=document.getElementById('root');

function fmt(ms){
  if(!ms||ms<=0)return'—';
  const s=Math.floor(ms/1000);
  if(s<60)return s+'s';
  const m=Math.floor(s/60),r=s%60;
  if(m<60)return m+'m '+r+'s';
  return Math.floor(m/60)+'h '+(m%60)+'m';
}
function dt(ts){
  if(!ts)return'—';
  try{return new Date(ts).toLocaleString();}catch(_){return String(ts);}
}
function sl(s){
  if(!s)return'UNKNOWN';
  return(typeof s==='string'?s:s.name||String(s)).toUpperCase();
}
function esc(s){
  return String(s||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

function render(d){
  const{name,status,stats={},start_time,end_time,environment,tags=[],launch_id,report_web_url}=d;
  const total=+(stats.total||0),passed=+(stats.passed||0);
  const failed=+(stats.failed||0),broken=+(stats.broken||0);
  const done=passed+failed+broken,pend=Math.max(0,total-done);
  const pct=total>0?Math.round(done/total*100):0;
  const sLabel=sl(status);
  const isRun=sLabel==='RUNNING';
  const dur=(start_time&&end_time&&end_time>start_time)?fmt(end_time-start_time):(isRun?'Running…':'—');
  const tagStr=(tags||[]).map(t=>esc(t.name)).join(', ')||'—';
  const env=esc(environment||'—');
  const lid=+launch_id;

  root.innerHTML=` + "`" + `<div class="card">
<div class="header">
  <div class="launch-name">🧪 ${esc(name||'Launch #'+lid)}</div>
  <span class="badge ${esc(sLabel)}">● ${esc(sLabel)}</span>
</div>
<div class="prog">
  <div class="prog-row"><span>${done}/${total} done</span><span>${pct}%</span></div>
  <div class="track"><div class="fill${isRun?' running':''}" style="width:${pct}%"></div></div>
</div>
<div class="stats">
  <div class="stat"><div class="sv p">${passed}</div><div class="sl">Passed</div></div>
  <div class="stat"><div class="sv f">${failed}</div><div class="sl">Failed</div></div>
  <div class="stat"><div class="sv b">${broken}</div><div class="sl">Broken</div></div>
  <div class="stat"><div class="sv s">${pend}</div><div class="sl">Pending</div></div>
</div>
<div class="meta">
  <span>⏱ ${dur}</span>
  <span>🏷 ${tagStr}</span>
  <span>🖥 ${env}</span>
  ${start_time?'<span>📅 '+dt(start_time)+'</span>':''}
</div>
<div class="actions">
  <button class="btn primary" onclick="ask('List test results for launch ${lid}')">📊 Results</button>
  <button class="btn" onclick="ask('Show failed tests in launch ${lid}')">✗ Failures</button>
  ${isRun
    ? '<button class="btn danger" onclick="ask(\'Close launch '+lid+'\')">■ Close</button>'
    : '<button class="btn danger" onclick="ask(\'Reopen launch '+lid+'\')">↩ Reopen</button>'}
  ${report_web_url?'<button class="btn" onclick="openUrl(\''+esc(report_web_url)+'\')">🔗 Report</button>':''}
</div>
</div>` + "`" + `;
}

(async()=>{
  if(!globalThis.ExtApps||!globalThis.ExtApps.App){
    root.innerHTML='<div class="error">ExtApps SDK not available</div>';return;
  }
  const app=new App('launch-dashboard',{},{});

  const applyTheme=ctx=>{
    if(ctx&&ctx.colorScheme==='dark')document.documentElement.classList.add('dark');
    else if(ctx&&ctx.theme==='dark')document.documentElement.classList.add('dark');
    else document.documentElement.classList.remove('dark');
  };
  app.onhostcontextchanged=applyTheme;

  app.ontoolresult=({content})=>{
    try{
      const raw=Array.isArray(content)?content[0]?.text:content;
      const data=typeof raw==='string'?JSON.parse(raw):raw;
      render(data);
    }catch(e){
      root.innerHTML='<div class="error">Parse error: '+String(e.message||e)+'</div>';
    }
  };

  window.ask=msg=>app.sendMessage({role:'user',content:[{type:'text',text:msg}]});
  window.openUrl=url=>app.openLink?.({url});

  await app.connect();
  applyTheme(app.getHostContext?.());
})();
</script>
</body>
</html>`

// ---------------------------------------------------------------------------
// Action Picker Widget HTML template
// ---------------------------------------------------------------------------

const actionPickerTemplate = `<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>
:root{--bg:#fff;--card:#f8f9fa;--text:#1a1a1a;--sub:#6b7280;--border:#e5e7eb;--accent:#2563eb}
html.dark{--bg:#1c1c1e;--card:#2c2c2e;--text:#f0f0f0;--sub:#9ca3af;--border:#3a3a3c;--accent:#3b82f6}
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:var(--bg);color:var(--text);padding:12px;font-size:13px;line-height:1.5}
.search{display:flex;gap:8px;margin-bottom:12px}.search input{flex:1;padding:8px 12px;border:1px solid var(--border);border-radius:8px;background:var(--card);color:var(--text);font-size:13px}
.list{display:flex;flex-direction:column;gap:8px;max-height:400px;overflow-y:auto}
.item{padding:12px;border:1px solid var(--border);border-radius:8px;background:var(--card);cursor:pointer;transition:.15s}
.item:hover{background:var(--accent);color:#fff;border-color:var(--accent)}.item-title{font-weight:600;margin-bottom:4px}
.item-method{display:inline-block;padding:2px 8px;border-radius:4px;font-size:11px;font-weight:700;background:var(--bg);margin-right:6px}
.item:hover .item-method{color:inherit}.empty{padding:32px;text-align:center;color:var(--sub)}
</style></head><body>
<div class="search"><input id="filter" type="text" placeholder="Filter results..."></div>
<div id="list" class="list"></div>
<script>
/*__EXT_APPS_BUNDLE__*/
const {App}=globalThis.ExtApps;
const filterInput=document.getElementById('filter');
const listDiv=document.getElementById('list');
let items=[];
function esc(s){return String(s||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');}
function render(){
  const q=filterInput.value.toLowerCase();
  const filtered=items.filter(it=>it.summary.toLowerCase().includes(q)||it.description.toLowerCase().includes(q)||it.path.toLowerCase().includes(q));
  listDiv.innerHTML='';
  if(filtered.length===0){listDiv.innerHTML='<div class="empty">No operations match your filter</div>';return;}
  for(const op of filtered){
    const el=document.createElement('div');
    el.className='item';
    const methodClass={GET:'#10b981',POST:'#3b82f6',PUT:'#f59e0b',DELETE:'#ef4444'}[op.method]||'#6b7280';
    el.innerHTML='<div class="item-title">'+esc(op.summary)+'</div><span class="item-method" style="background:'+methodClass+';color:#fff">'+esc(op.method)+'</span><span style="color:var(--sub)">'+esc(op.path)+'</span>';
    el.onclick=()=>app.sendMessage({role:'user',content:[{type:'text',text:'Select operation: '+op.operation_id}]});
    listDiv.append(el);
  }
}
(async()=>{const app=new App('action-picker',{},{});
app.ontoolresult=({content})=>{
  try{const raw=Array.isArray(content)?content[0]?.text:content;const data=typeof raw==='string'?JSON.parse(raw):raw;items=data.results||[];filterInput.value='';render();}
  catch(e){listDiv.innerHTML='<div class="empty">Error: '+String(e)+'</div>';}
};
filterInput.addEventListener('input',render);await app.connect();})();
</script></body></html>`

// ---------------------------------------------------------------------------
// Results Display Widget HTML template
// ---------------------------------------------------------------------------

const resultsDisplayTemplate = `<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>
:root{--bg:#fff;--card:#f8f9fa;--text:#1a1a1a;--sub:#6b7280;--border:#e5e7eb;--success:#16a34a;--error:#dc2626}
html.dark{--bg:#1c1c1e;--card:#2c2c2e;--text:#f0f0f0;--sub:#9ca3af;--border:#3a3a3c;--success:#22c55e;--error:#ef4444}
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:var(--bg);color:var(--text);padding:12px;font-size:13px;line-height:1.5}
.header{padding:12px;background:var(--card);border:1px solid var(--border);border-radius:8px;margin-bottom:12px;display:flex;align-items:center;gap:8px}
.status-ok{color:var(--success)}.status-err{color:var(--error)}
.body{padding:12px;background:var(--card);border:1px solid var(--border);border-radius:8px;font-family:monospace;font-size:11px;max-height:300px;overflow-y:auto;white-space:pre-wrap;word-break:break-word}
.empty{padding:32px;text-align:center;color:var(--sub)}
</style></head><body>
<div id="root"><div class="empty">Executing...</div></div>
<script>
/*__EXT_APPS_BUNDLE__*/
const {App}=globalThis.ExtApps;
const root=document.getElementById('root');
function esc(s){return String(s||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');}
function formatJSON(obj){try{return JSON.stringify(obj,null,2);}catch(e){return String(obj);}}
function render(data){
  const isErr=data.isError;const status=isErr?'Error':'Success';const statusClass=isErr?'status-err':'status-ok';
  let content='';
  if(Array.isArray(data.content)&&data.content.length>0){const txt=data.content[0]?.text||'';content=txt;}
  try{const parsed=JSON.parse(content);content=formatJSON(parsed);}catch(_){}
  root.innerHTML='<div class="header"><span class="'+statusClass+'">● '+status+'</span></div><div class="body">'+esc(content)+'</div>';
}
(async()=>{const app=new App('results-display',{},{});
app.ontoolresult=(data)=>{render(data);};
await app.connect();})();
</script></body></html>`

// ---------------------------------------------------------------------------
// Widget registration
// ---------------------------------------------------------------------------

const quickstartContent = `# TestOps MCP Server

Connect to Allure TestOps to manage test launches, results, and test cases directly from your AI assistant.

## Quick Start

1. Configure your Allure token via ` + "`configure_allure_token`" + ` or the ` + "`ALLURE_TOKEN`" + ` environment variable
2. Use ` + "`list_projects`" + ` to see available projects
3. Use ` + "`list_launches`" + ` to browse test runs
4. Use ` + "`get_launch_dashboard`" + ` for a visual launch overview with live stats

## Tool Groups

- **Launches** — list, get, create, close, reopen test launches
- **Results** — list test results, get details, update statuses
- **Test Cases** — search, create, update test cases and steps
- **Analytics** — trends, flaky tests, defect distribution
- **Bulk** — bulk status updates across test results and test cases
- **Search** — ` + "`search_testops_operations`" + ` discovers all 300+ API operations

## Prompts

Use the built-in prompts for common workflows:
- ` + "`analyze-test-failures`" + ` — deep-dive into failures in a specific launch
- ` + "`launch-report-summary`" + ` — generate an executive summary for a launch
`

// registerWidgets registers all MCP app tools and their associated resources.
// Called once from NewRegistry after all other tools are registered.
func (r *Registry) registerWidgets() {
	r.RegisterResource(&Resource{
		URI:      "allure://docs/quickstart",
		Name:     "TestOps MCP Quickstart",
		MimeType: "text/markdown",
		GetHTML:  func() string { return quickstartContent },
	})

	if r.allure == nil {
		return
	}

	// get_launch_dashboard — visual launch dashboard widget
	r.register(&Tool{
		Name: "get_launch_dashboard",
		Description: "Get an interactive visual dashboard for a launch showing real-time status, " +
			"progress bar, and pass/fail statistics. " +
			"Renders an inline widget in Claude Desktop and claude.ai. " +
			"Use instead of (or in addition to) get_launch_report for a richer view.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
			},
			"required": []string{"launch_id"},
		},
		Meta: map[string]any{
			"ui": map[string]any{
				"resourceUri": launchDashboardURI,
			},
		},
		Handler: Typed(r.getLaunchDashboard),
	})

	// Register the widget resource — HTML is rendered lazily on first request
	// so the bundle fetch doesn't block server startup.
	r.RegisterResource(&Resource{
		URI:      launchDashboardURI,
		Name:     launchDashboardName,
		MimeType: widgetMimeType,
		GetHTML: func() string {
			bundle := getExtAppsBundle(r.logger)
			return strings.ReplaceAll(launchDashboardTemplate, "/*__EXT_APPS_BUNDLE__*/", bundle)
		},
	})

	// Register action picker widget resource
	r.RegisterResource(&Resource{
		URI:      "ui://widgets/action-picker",
		Name:     "Action Picker",
		MimeType: widgetMimeType,
		GetHTML: func() string {
			bundle := getExtAppsBundle(r.logger)
			return strings.ReplaceAll(actionPickerTemplate, "/*__EXT_APPS_BUNDLE__*/", bundle)
		},
	})

	// Register results display widget resource
	r.RegisterResource(&Resource{
		URI:      "ui://widgets/results-display",
		Name:     "Results Display",
		MimeType: widgetMimeType,
		GetHTML: func() string {
			bundle := getExtAppsBundle(r.logger)
			return strings.ReplaceAll(resultsDisplayTemplate, "/*__EXT_APPS_BUNDLE__*/", bundle)
		},
	})
}

// ---------------------------------------------------------------------------
// Resource watch / subscriptions
// ---------------------------------------------------------------------------

// watchLaunch polls the launch status every 10 s and calls publishResource when it changes.
// It runs until ctx is cancelled.
func (r *Registry) watchLaunch(ctx context.Context, launchID int64) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	var lastStatus string
	uri := launchDashboardURIFor(launchID)
	for {
		select {
		case <-ticker.C:
			if r.publishResource == nil || r.allure == nil {
				continue
			}
			details, err := r.allure.GetLaunchDetails(ctx, launchID)
			if err != nil {
				continue
			}
			currentStatus := fmt.Sprintf("%v", details.Status)
			if currentStatus != lastStatus {
				lastStatus = currentStatus
				r.publishResource(uri)
			}
		case <-ctx.Done():
			return
		}
	}
}

// StartLaunchWatch begins polling a launch and publishing updates to subscribers.
// Call this after a client subscribes to a launch dashboard resource.
func (r *Registry) StartLaunchWatch(ctx context.Context, launchID int64) {
	go r.watchLaunch(ctx, launchID)
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

type getLaunchDashboardArgs struct {
	LaunchID int64 `json:"launch_id"`
}

func (r *Registry) getLaunchDashboard(ctx context.Context, args getLaunchDashboardArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("fetching launch dashboard", map[string]any{"launch_id": args.LaunchID})

	details, err := r.allure.GetLaunchDetails(ctx, args.LaunchID)
	if err != nil {
		r.logger.Error("get launch details", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("get launch details: %w", err)
	}

	stats, err := r.allure.GetLaunchStatistics(ctx, args.LaunchID)
	if err != nil {
		r.logger.Error("get launch statistics", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("get launch statistics: %w", err)
	}

	tags := make([]map[string]any, len(details.Tags))
	for i, tag := range details.Tags {
		tags[i] = map[string]any{"id": tag.ID, "name": tag.Name}
	}

	return map[string]any{
		"launch_id":      details.ID,
		"name":           details.Name,
		"status":         details.Status,
		"start_time":     details.StartTime,
		"end_time":       details.EndTime,
		"environment":    details.Environment,
		"tags":           tags,
		"report_web_url": details.ReportWebUrl,
		"stats": map[string]any{
			"total":  stats.Total,
			"passed": stats.Passed,
			"failed": stats.Failed,
			"broken": stats.Broken,
		},
	}, nil
}
