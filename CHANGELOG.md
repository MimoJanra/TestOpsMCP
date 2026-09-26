# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- **All test-case bulk operations moved from v1 to v2 endpoints** (`/api/testcase/bulk/*` → `/api/v2/test-case/bulk/*`): clone, move, remove, mute, status, layer, tag add/remove, member add/remove, issue add/remove, external links, test plan create, run new/existing (also used by `run_test_case`). v1 has repeatedly returned success while doing nothing or 500'd in this project's Allure instance; each v2 endpoint was verified live on project 408 before switching. Tool parameters are unchanged — only the selection shape differs (`testCasesInclude` instead of `leafsInclude`).
- **Folder tools rebuilt on the v2 tree API — parameters changed.** In Allure a folder is a custom field value that only exists within a named tree (e.g. "Suites", "Features"), and the v1 path-based endpoints never passed a tree, which is why `create_test_case_folder` failed with "tree has no group at level 0". `browse_test_case_tree`, `get_test_case_tree_folders`, `create_test_case_folder` and `move_test_cases_to_folder` now take `tree_id` (optional when the project has exactly one tree) and node ids (`parent_node_id` / `node_id`) instead of `path` / `parent_path` / `dest_path`. `move_test_cases_to_folder` resolves the API's tree leaf ids (which change on every move) from test case ids itself. All confirmed live.
- **`update_test_case` `manual_scenario` now replaces the scenario through the step API** — it used to store every step body as `<empty>` (confirmed live). The scenario is now rebuilt step by step with real text and expected results, in a new nested shape `{"steps": [{"body", "expected_result", "steps": [...]}]}`; the legacy flat `{"type": "body"|"expected", "body"}` entries are still accepted. `""` now clears a text field and `[]` clears tags/members/links — both used to be silently ignored.
- **`add_test_case_members` is now additive** (v2 `bulk/member/add`) — it used to replace all of the test case's members. `role` is optional: the project's role scheme decides the stored role (confirmed live). `get_test_case_members` now returns each member's role.
- **`remove_test_case_members`: `project_id` is optional** (looked up from the test case), and a `project_id` that doesn't match the test case's project is refused — it used to be a silent no-op.
- **`move_test_case_step` requires a position** (`after_id`, `before_id` or `parent_id`) — without one the API silently un-nested the step. `move_test_case_step` and `copy_test_case_step` gained `test_case_id`, which resolves the anchor's parent (the API 404'd on a nested anchor); `copy_test_case_step` returns the new `step_id`.
- **`bulk_run_test_cases_new_launch`: `launch_name` is now required** (the API requires it) and the task result includes the new `launch_id`.
- **`bulk_clone_test_cases`** gained `name_suffix` and `ignore_tags`. Its description no longer claims to return the new ids: the API clones asynchronously (202, no body).
- **`bulk_create_test_plan`** returns `test_plan_id`, `name` and `test_cases_count`.
- **`get_launch_status` returns `OPEN`/`CLOSED` plus the per-status `statistic`** — it was always null, because the launch DTO has no status field. `get_launch_details` and `list_launches` report `status` as `OPEN`/`CLOSED` with a `closed` flag and no longer return `start_time`/`end_time`/`environment`/`description`/`report_web_url`, which the API never sends; `get_launch_details` adds `autoclose`, `external`, `links`, `issues` and `statistic`. `get_launch_report` now includes `skipped` and `unknown`.
- **`get_test_result`** adds `tested_by`, `hidden`, `manual`, `category`, `layer` and `links`.
- **`get_test_case_attachments`**: `size` is the real content length, `created_date` removed (not in the API), `missed` added.
- **`search_testops_operations` matching** — token-based (multi-word intents used to return nothing), verbs hint the HTTP method, v2 paths and exact path segments rank higher, and an exact operation id returns that operation.
- **Message size limit raised from 1 MiB to 32 MiB** on stdio and HTTP, so a 20 MiB attachment fits as base64. A reverse proxy in front of an HTTP deployment must allow the same (`client_max_body_size 32m`); see `docs/DEPLOYMENT.md`.

### Added

- **`upload_test_case_attachment`** — upload a file to a test case and optionally show it in a step or its expected result (the web UI's "File" block). Takes `file_path` or `content_base64` + `file_name`. `file_path` is only allowed in local stdio mode: it reads the server's disk, and on a shared HTTP server a client could otherwise pick the "stdio" session id via its `Mcp-Session-Id` header and exfiltrate server files — so the permission is set from the transport at startup, not from the request. Max file size 20 MiB, so the base64 form fits in one 32 MiB MCP message.
- **`add_test_case_step_table`** — add a table to a step (the web UI's "Table" block). Confirmed live that a step table is a CSV attachment shown in the step as a child node, not a table in the step's rich-text body, so the tool uploads the rows as CSV and attaches it.
- **`list_test_case_trees`** — lists a project's test case trees, to get the `tree_id` the folder tools need.
- **`get_test_case_attachment_content`** — download an attachment: text as `content`, binary as `content_base64`. `save_to` writes a local file instead — stdio mode only, and it never overwrites an existing file (`O_EXCL`). Max 20 MiB.
- **`rename_test_case_folder`** — renames a folder in place via v2 `renameGroup`; the node keeps its id and its test cases follow. Confirmed live.
- **`delete_test_case_folder`** — unassigns the folder's test cases and removes the folder. The API's `deleteGroup` only unassigns (the empty folder stays), so the tool also deletes the folder's backing custom field value — unless another same-named folder shares it, in which case the status is `emptied`. Confirmed live.
- **Test plan tools: `list_test_plans`, `get_test_plan`, `run_test_plan`, `rename_test_plan`, `delete_test_plan`** — plans could only be created (`bulk_create_test_plan`) or added to a launch. `run_test_plan` starts a new launch and returns its `launch_id`; `delete_test_plan` asks for confirmation. Confirmed live.
- **`delete_launch`** — permanently deletes a launch and its results; asks for confirmation via elicitation.
- **Defect tools: `list_defects`, `get_defect`, `create_defect`, `update_defect`, `delete_defect`** — project-level defects (open/closed filter, close/reopen via `update_defect` `closed`). `delete_defect` asks for confirmation. Confirmed live.
- **`create_test_case` accepts the whole test case in one call** — `precondition`, `expected_result`, `full_name`, `automated`, `status_id` + `workflow_id` (must be passed together), `test_layer_id`, `tags`, `links`, `members`, `custom_fields` (`[{custom_field_id, value_id | name}]`, a name is found or created) and `steps` (`[{body, expected_result, steps}]`). Returns `steps_created`. Confirmed live.
- **`create_test_case_step` gained `expected_result` and `before_id`**; the parent of a nested `after_id`/`before_id` anchor is resolved automatically.
- **`detach_test_case_automation` gained `use_scenario_from_test_result`** (turn the latest automated run's steps into the manual scenario).
- **`resolve_test_result`/`bulk_resolve_test_results` gained `category_id` and `message` parameters** — the web UI's "Change status" dialog has Status/Category/Details fields, but the tools only ever sent `status`, silently dropping any category or reason text a caller tried to attach. Both fields were already present and optional in the real request schemas (`ResolveRequestV2Dto`/`TestResultBulkResolveDto`); the tools just never exposed them. Both tools now go through the v2 endpoint (`/api/v2/test-result/bulk/resolve`) — confirmed live that the v1 endpoints (`/api/testresult/{id}/resolve`, `/api/testresult/bulk/resolve`) accept `message` and return success but silently drop it; the single-result tool looks up the result's launch and sends a one-item selection.

### Fixed

- **`bulk_add_test_case_issues` could link every existing issue in the instance to a test case** — confirmed live: an issue sent without `integration_id`/key isn't rejected by the API, it attaches all issues (113 real tickets on one sandbox test case, since cleaned up; no Jira-side links were created). The tool now requires `integration_id` and `display_name` for every issue and refuses the call otherwise.
- **`bulk_set_test_case_layer` rejected every built-in layer** with `layer_id must be positive` — built-in layer ids are negative (Unit/UI/API Tests = -1/-2/-3), same as the earlier `workflow_id` fix.

- **`list_test_results`'s `status` filter silently did nothing** — reported live: a filtered request returned every status regardless of the value passed. `GET /api/testresult` has no `status` query parameter at all (only `launchId`/`page`/`size`/`sort` per the spec); sending one is just ignored. Now scans the launch's results client-side and paginates over the matches when `status` is set (capped at 20,000 scanned results; a `truncated` flag surfaces if the cap is hit). `analyze_launch_failures` had the exact same bug — it silently analyzed the first N results of *any* status, not the actual failures — and is fixed the same way.
- **`list_test_results` pagination could skip or duplicate results at page boundaries** — the request only sorted by `createdDate,DESC` with no tiebreaker, so results sharing the same timestamp had no guaranteed stable order across separate page requests. Added an explicit `id,ASC` secondary sort.
- **`list_test_results`'s `size` silently clamped to 100** with no indication in the tool description — confirmed live the API itself accepts far larger pages (`size=300` returns a full 300-item page, no server-side clamp). Raised the cap to 1000 and documented it.
- **`bulk_mute_test_results` 500'd** while the single-result `mute_test_result` worked fine — same NOT NULL constraint on the mute reason's `name` as `mute_test_result`/`bulk_mute_test_cases` before their fixes, just missed in that earlier pass (`TestResultBulkMuteDto` marks `name` optional in the spec, but omitting it crashes the DB insert). Reported live. Now always sends `name` (defaulting to the reason text, or `"Muted via MCP"` if no reason given).
- **The stdio server exited on a JSON-RPC line over the size limit** (`bufio.Scanner: token too long`). An oversized line is now answered with a JSON-RPC error and the server keeps serving.
- **Tool arguments are validated against the schema's `required` lists** (top level and inside array items) before the handler runs — e.g. `set_test_case_relations` without `relations` used to wipe all relations.
- **Project `code` was always empty** — the API field is `abbr`. `find_project` by code now works. Confirmed live.
- **`set_test_case_issues` and `bulk_add_test_case_issues` 500'd on every call** — the issue key is now sent as `name`. `set_test_case_issues` requires `integration_id` and `display_name` per issue. `get_test_case_issues` returned an empty `display_name` (now read from `name`) and adds `summary`/`status`. Confirmed live.
- **`list_test_cases`/`search_test_cases` always returned `automation_status: null`** — now `automated`/`manual`. `search_test_cases` returned `project_id: 0`. Its description now says values are case-sensitive (`status = "active"` matched nothing).
- **Setting a step's expected result flattened the body's rich-text formatting** — the stored `bodyJson` is sent back now. It also no longer overwrites a file/table node that is first in the expected-result container, and editing a file/table node's body is refused (it silently dropped the attachment). Confirmed live.
- **`detach_test_case_automation` failed when `status_id`/`workflow_id` were omitted** — they were sent as 0; now left out.
- **A mute reason over 255 characters 500'd** (`mute_test_result`, `bulk_mute_test_results`, `bulk_mute_test_cases`) — the mute name column is limited to 255; the name is truncated and the reason kept whole. Confirmed live.
- **`assign_test_result` reported success when Allure ignored the assignment** (already resolved results) — it now reads the assignee back and errors. Descriptions of `bulk_assign_test_results`/`bulk_mute_test_results` document round-robin assignment, unassign-on-omit and skipped not-run results.
- **`bulk_unmute_test_results` left the test cases in the project's muted list** — it now also removes their mute records. Confirmed live.
- **`close_launch`/`reopen_launch` on a nonexistent launch reported success** — the launch is now looked up first.
- **`merge_launches` with the target among the sources deleted the launch** — self-merges and duplicate ids are refused, sources already merged are reported on partial failure, and the tool is annotated destructive.
- **Built-in custom fields (negative ids: Epic/Feature/Story/Suite/Component) were rejected** by `get_custom_field`, `get_project_custom_field`, `update_project_custom_field`, `create_custom_field_value` and `list_custom_field_values`. `get_project_custom_field` on an unattached field failed with `decode response: EOF` — now a clear error. `add_custom_fields_to_project` reports ids that weren't attached (the API ignores unknown ids). Descriptions corrected: `locked` doesn't block edits, renaming a value gives it a new id (test cases follow), `global` is irreversible, and delete/detach prerequisites are listed. Confirmed live.
- **`move_test_cases_to_folder`/`create_test_case_folder` silently stripped folders** when `node_id`/`parent_node_id` was a leaf id, an unknown id or another tree's folder — both now validate it's a folder of the tree. Tree listing sort has tiebreakers (name, createdDate, testCaseId), so paging no longer duplicates or skips same-named items. `get_test_case_tree_folders` returned empty pages — it now collects all folders of the level before paginating and returns `total`. Tree tools return each folder's `custom_field_value_id`.
- **`set_test_case_keys` accepted a key without `integration_id`/`name`** that the API silently dropped — now refused.
- **Tag descriptions claimed a new tag name 409s** — a name-only tag is created.
- **`get_test_case_examples` returned at most 10 rows** — now all rows.
- **`analyze_launch_failures` skipped broken results** and succeeded on a nonexistent launch — both fixed.
- **`execute_testops_operation` sent large numbers as `1.78e+12`** in query/path params, sent array query params once instead of repeating them (`?id=1&id=2`), and mangled binary responses — binary is now returned as `content_base64`.
- **Sampling/elicitation requests to a client that didn't declare the capability blocked until timeout** (`analyze_launch_failures` waited 120 s) — the server now tracks client capabilities from `initialize` and fails immediately.
- **A task cancelled with `cancel_task` flipped to succeeded/failed** when the work finished later — it stays cancelled and the late outcome goes into its message.
- **`CustomFieldValueWithCfDto` model** fixed to the spec shape (`{id | name, customField: {id}}`).

## [2.4.0] - 2026-09-22 - Fix Bulk Operation DTOs, Test Case Relations, Examples Decode, and Member Roles

### Added

- **`update_launch`** — there was no way to rename a launch or change its `autoClose`/`external` flags after creation. Exposes `PATCH /api/launch/{id}`. Confirmed live: renamed an existing launch and verified via `get_launch_details`.

### Fixed

- **`mute_test_result` failed with `500 null value in column "name" of relation "mute"`** on every call — the request only sent `reason`, but the API's `TestResultMuteReason` schema requires `name` too (a DB NOT NULL constraint the OpenAPI spec doesn't mark required). Now always sends `name` (defaulting to the reason text, or `"Muted via MCP"` if no reason given). Confirmed live.
- **`copy_launch` failed with `500 null value in column "name" of relation "launch"`** on every call — the copy request body was sent empty, but the endpoint requires `launchName`. Added an optional `launch_name` parameter; when omitted, the tool looks up the original launch's name and defaults to `"<original name> (copy)"`. Confirmed live.
- **`run_test_case` 409'd with `{"field":"selection","must not be null"}`** even when the test case was already in the launch — the request only sent `testCaseIds`/`launchId`, but the endpoint requires a full `TestCaseTreeSelectionDto` including the test case's `projectId`. The tool now looks up the test case's project automatically, so its own parameters (`test_case_id`, `launch_id`) are unchanged. Confirmed live.
- **`get_project_stats`, `list_test_cases`, and `list_deleted_test_cases` always reported `project_id: 0`** in their results — each read `projectId` back off the API response item, but neither `/api/project/{id}/stats` nor the test-case list endpoints actually include that field per item (it's implied by the request, not echoed back). All three now return the `project_id` that was actually requested. Confirmed live for `get_project_stats` and `list_test_cases`.
- **`delete_test_case` and `delete_test_case_version` reported "unexpected status 202"** even though the delete actually succeeded — the API returns `202 Accepted` for these (an async delete), which wasn't in the client's accepted-status list.
- **`clone_test_case` failed with `400 Required request body is missing`** on every call — the endpoint requires a request body even though every field within it is optional. Now sends `{}` instead of no body at all.
- **`bulk_set_test_case_status` rejected every real workflow ID** with `workflow_id must be positive` — workflow IDs are commonly negative (e.g. `-1` for "Default Manual"). Relaxed to `workflow_id != 0`, mirroring the earlier `status_id` fix.
- **`bulk_remove_test_case_tags` and `bulk_remove_test_case_members` reported success but didn't detach anything.** Both sent the *add* endpoints' DTO shape (`{tags/members: [...], selection}`) to the *remove* endpoints, which actually expect `{ids: [...], selection}` — tag/member ids to detach, not the full entity payload. `bulk_remove_test_case_tags` now resolves tag names to their (global, project-wide) ids by looking them up on the target test cases first.
- **`bulk_mute_test_cases` 500'd** — the request only sent `selection`, but the endpoint requires a `mute: {name, reason}` object (the DB has a NOT NULL constraint on `mute.name` the spec doesn't surface, same class of bug as the `mute_test_result` fix). Added an optional `reason` parameter; defaults to `"Muted via MCP"` when omitted.
- **`delete_test_case_external_link` silently no-op'd when removing a test case's last remaining link** — the shared PATCH DTO's `Links []ExternalLinkDto` field used `omitempty`, so clearing to an empty list serialized identically to "don't touch links" (the field was dropped from the request entirely) and the server kept the old links. `Links` is now `*[]ExternalLinkDto`, which lets an explicit empty slice still serialize as `"links": []`.
- **`set_test_case_relations` sent an invalid request** — `RelationDto` was missing the required `type` field entirely (and the tool's description advertised nonexistent relation types `blocks`/`is blocked by`). Added `type` (validated against the real enum: `related to`, `clones`, `is cloned by`, `duplicates`, `is duplicated by`, `automates`, `is automated by`) and corrected the docs.
- **`get_test_case_examples` failed with `decode response: json: cannot unmarshal object into Go value of type [][]allure.TestCaseExampleParam`** on every call — the GET endpoint returns a paginated page object (`{content: [...], ...}`), not a bare array of rows (that shape is only the POST request body). Now decodes the page and returns each row's `id`, `status`, and `parameters`.
- **`list_muted_test_cases` always reported `project_id: 0`** — same class of bug as `get_project_stats`/`list_test_cases`/`list_deleted_test_cases`, just missed in that pass. Now returns the requested `project_id`.
- **`add_test_case_members`/`bulk_add_test_case_members` failed with a misleading `400 Some role users not found`** — each member needs a `role` (the API's `MemberDto` requires it), but neither tool's schema advertised the field, so callers had no way to know to send it. Documented the requirement (with how to look up valid role ids via `GET /api/role`, e.g. `-1` "Owner"/`-2` "Lead") and added explicit validation with an actionable error instead of the confusing 400. Also documented that the member id must be an existing project collaborator (via `GET /api/member/suggest?projectId=`), not just any org-wide user id. Confirmed live: added a project collaborator as an "Owner" member.

## [2.3.1] - 2026-09-22 - Fix AQL Syntax and Remaining Custom Field Value/Field-ID Confusion

### Fixed

- **`search_test_cases`/`validate_test_case_query`'s documented AQL syntax was backwards.** The tool description claimed string literals must be single-quoted (`status = 'active'`) and double quotes cause a 400 — the actual Allure AQL grammar (per docs.qameta.io/allure-testops/advanced/aql/) requires **double**-quoted string literals, uses `~=` for partial match (not `~` alone), and references custom fields via `cf["Name"] = "value"` bracket notation. Every example in the old description, including the tool's own advertised syntax, was invalid AQL. Reported live: a query returning `valid: false` for every string-literal query tried, including the tool's own documented example. Corrected the description and all doc examples to the real, double-quoted syntax.
- **`get_test_case_custom_fields` still returned empty `values` even with the `projectId` fix from 2.2.2** — the dedicated `GET /api/testcase/{id}/cfv` endpoint is unreliable on this API regardless of that param, confirmed by comparing against `get_test_case`'s own (correct) `customFields` field. Rewrote the tool to source from the test case's overview instead of that endpoint at all.
- **`bulk_remove_test_case_custom_fields` still didn't remove anything after the 2.2.2 v1→v2 endpoint switch.** The v2 endpoint's `ids` parameter means cfv **value** ids, not custom field ids — passing a field id is a silent no-op (204, nothing removed). Confirmed live by calling the raw endpoint with the value id directly, which worked. The tool now resolves each test case's current value id(s) for the given field(s) via its overview before removing. `update_test_case_custom_fields`'s clear-then-set implementation, and its rollback snapshot (which was silently reading the same broken `/cfv` GET and would never have restored a real value), are fixed the same way.

## [2.3.0] - 2026-09-22 - Add Custom Field Management, Fix update_test_case_step Body Drop

### Added

- **13 new custom field management tools**, covering the three layers the API exposes beyond per-test-case values: field **definitions** (`create_custom_field`, `get_custom_field`, `update_custom_field`, `delete_custom_field`, `set_custom_field_archived`), **project attachment** (`list_project_custom_fields`, `get_project_custom_field`, `add_custom_fields_to_project`, `remove_custom_field_from_project`, `update_project_custom_field`), and a project's **value catalog** (`create_custom_field_value`, `update_custom_field_value`, `delete_custom_field_value`). Previously the server could only read and set values already defined on a field — there was no way to create a new custom field, attach it to a project, or add/rename/remove its selectable value options.

### Fixed

- **`update_test_case_custom_fields` could permanently wipe a field's values on a partial failure.** Its clear-then-set implementation (see 2.2.2) had no rollback: if the clear succeeded but the set failed (transient error, bad value on one field), the field was left empty with no way to recover the original value. The tool now snapshots each touched field's current values first and, on any failure, re-clears the affected fields (in case the failed call partially applied rows — the bulk endpoints aren't atomic across rows) and restores whichever ones previously had a value, on a best-effort basis; the returned error states whether the restore succeeded.
- **`bulk_add_test_case_custom_fields` silently reported success while adding nothing** when a field's `values` list was empty (e.g. a caller bug that forgot to populate it) — this tool only adds values, so an empty list is now rejected with a clear error instead of being sent to the API as a no-op.
- **`update_test_case_step` silently dropped `body` when `expected_result` was also set on a step that already had an expected result** — reported `{"status":"updated"}`, but only `expected_result` was actually saved. The branch that PATCHes the parent step's body only ran on a step's *first* expected result (when its expected-result container didn't exist yet); on every subsequent edit, body was never sent anywhere. Confirmed live: [#20](https://github.com/MimoJanra/TestOpsMCP/issues/20).
- **`test_case_id` is now unconditionally required for `update_test_case_step`** (was only "recommended" for a body-only edit) — without it the tool can't detect an existing expected result and silently wipes it on a body-only edit, which is exactly what happened while working around the bug above.

## [2.2.2] - 2026-09-13 - Fix Custom Field Bulk Remove and Single-Case Update

### Fixed

- **`bulk_remove_test_case_custom_fields` reported success but silently left the custom field value in place.** The v1 endpoint (`/api/testcase/bulk/cfv/remove`), unlike the v2 one, doesn't actually clear the value on this API's backend even though it returns 204. Switched to `/api/v2/test-case/bulk/cfv/remove`, mirroring the earlier `bulk_add_test_case_custom_fields` v1→v2 fix. Confirmed live: [#18](https://github.com/MimoJanra/TestOpsMCP/issues/18).
- **`get_test_case_custom_fields` returned stale or empty values for some custom fields** (e.g. a multi-select field that `get_test_case`'s full response showed correctly) — the underlying `GET /api/testcase/{id}/cfv` endpoint requires a `projectId` query parameter per the API spec, which was never sent. The tool now looks up the test case's project via its overview and includes it.
- **`update_test_case_custom_fields` (single-case) failed with `500 An unexpected error occurred` on every call**, including an empty-array body on an unrelated test case — the dedicated `PATCH /api/testcase/{id}/cfv` endpoint is unconditionally broken on this API's backend, regardless of payload. Reimplemented on top of the (now-fixed) bulk v2 endpoints with a single test case ID: clear each named field first, then set the desired values — since bulk remove only supports clearing a whole field, not individual values, "update" here means replace, not merge.

### Changed

- **`get_test_case`, `get_test_case_steps`, and `create_test_case_step` now document a discovered Allure TestOps architecture gap**: a test case can store its steps in either the modern "manual scenario" tree (what these tools read/write, keyed by step IDs) or a legacy, ID-less `scenario` field returned alongside it in `get_test_case` — never both. `hasManualScenario: false` with a non-empty legacy `scenario.steps` means the case's real content lives only in the legacy field. Confirmed live (2026-08-27, tassta.testops.cloud project 170 case 13403): calling `create_test_case_step` on such a case immediately switches the web UI to showing only the new, near-empty modern tree — the legacy steps become invisible in the UI. There is no API path to migrate legacy steps automatically; the tool descriptions now instruct recreating all existing legacy steps (body + expected_result) in the same pass as any new addition, to avoid apparent data loss.

## [2.2.1] - 2026-08-26 - Fix Expected-Result Model in update_test_case_step

### Fixed

- **`update_test_case_step` was spawning spurious empty "Expected Result" placeholder child nodes on plain `body`-only edits**, not just when setting `expected_result` — the 2.2.0 fix for the opposite problem (losing an existing expected result on a body-only edit) unconditionally sent `withExpectedResult=true` on every call, which the API treats as "materialize expected-result bookkeeping for this step" regardless of whether `expected_result` was actually being set. Reported live with full repro in [#16](https://github.com/MimoJanra/TestOpsMCP/issues/16). `withExpectedResult` is now sent only when this call sets `expected_result`, or when `test_case_id` shows the step already has an expected result to preserve.
- **`update_test_case_step` set `expected_result` text that the web UI silently never displayed.** 2.2.0's "verify and repair" logic (`ensureExpectedResultText`) also made this worse: it didn't converge, and its own repair PATCH went through the same always-on client method, recursively spawning another nested placeholder off the node it had just written — removed entirely. Root cause, confirmed live against a real Allure TestOps instance (tassta.testops.cloud project 408) by diffing the step tree before/after typing into the web UI's own Expected Result field: `expectedResultId` points to a **container** step whose own `body` the UI does not read at all — the UI instead renders a list of the container's **child** steps (Allure supports multiple expected results per step; typing in the UI appends a new child rather than editing one). `update_test_case_step` now creates a child under the container when none exists, and replaces the first existing child's text on a later call (matching "set the expected result" semantics for the tool's single string parameter) instead of writing to the container itself.

## [2.2.0] - 2026-08-25 - Fix Async Status Codes, Custom Fields, and Stdio Confirmation Dialogs

### Added

- **`--version` CLI flag** — prints the build version and negotiated MCP protocol version, then exits (`testops-mcp v2.2.0 (MCP protocol 2025-11-25)`).
- **`create_test_tag` tool** — creates a new project-wide tag via `POST /api/tag` and returns its ID. There was previously no way to create a brand-new tag through the available tools.
- **`list_custom_field_values` tool** — lists the valid values defined for a custom field within a project (e.g. allowed Priority/Section options) via `GET /api/project/{id}/cfv`, so a value ID can be looked up instead of guessed.

### Fixed

- **`bulk_set_test_case_status` rejected legitimate workflow status IDs.** Allure TestOps uses negative IDs for built-in workflow statuses, but the tool validated `status_id` as `> 0`. Relaxed to `!= 0`.
- **Tags with `id: 0` couldn't be created via `set_test_case_tags`.** `TestTagDto.ID` had no `omitempty`, so a new tag (no ID yet) always serialized as `"id":0` instead of omitting the field — likely breaking the API's create-by-name path. Added `omitempty`.
- **`update_test_case_step` silently deleted a step's expected result.** `PATCH /api/testcase/step/{id}` requires the `withExpectedResult=true` query parameter (see `spec/testops.json`) or the API detaches `expectedResultId` and deletes the expected-result child node — even on a body-only edit. The client now always sends `withExpectedResult=true`.
- **`bulk_add_test_case_custom_fields` returned 500 for every custom field** (reproduced on `isAutomated`, `Priority`, `Section1`) even though the request matched the documented v1 schema for `POST /api/testcase/bulk/cfv/add`. Switched to the v2 endpoint (`POST /api/v2/test-case/bulk/cfv/add`), which uses a flat per-value shape (`{id, customField}` per row) instead of v1's field-with-nested-values shape.
- **`update_test_case_custom_fields` / `bulk_add_test_case_custom_fields` still 500'd even with a valid, existing value id** — the server error was `null value in column "name" of relation "custom_field_value" violates not-null constraint`, meaning it tries to INSERT a new value row instead of referencing the existing one when `name` is missing. Both tools now require `values: [{id, name}]` instead of bare `value_ids`, and validate that `name` is set before sending the request.
- **`assign_test_result`, `mute_test_result`, `resolve_test_result`, `unmute_test_result`, `copy_launch`, `restore_test_case_version`, `move_test_cases_to_folder` (drag-and-drop), and all `bulk_*` operations built on the shared `bulkPost` helper (e.g. `bulk_move_test_cases`) treated the API's actual success response (`202 Accepted`) as an error** — the calls succeeded server-side but were reported to the caller as "Tool execution failed". Added `202` to each of these client methods' accepted status codes.
- **`copy_launch` additionally never returned the new launch's ID** — the endpoint responds `202 Accepted` with no body per the API spec, so the client can't read one back; the tool result now says so explicitly instead of trying (and failing) to decode a launch object, and points the caller at `list_launches`.
- **`delete_test_case` / `bulk_delete_test_cases` always failed over stdio with "no interactive session is available"**, even with explicit user confirmation — the stdio transport never wired up the elicitation/sampling machinery (`sessctx.WithElicit`/`WithSampling`) that the HTTP transports use, and its single-threaded read loop couldn't have waited for a client reply mid-request anyway. The stdio transport now registers a persistent session (`Server.StdioSession`), dispatches each request on its own goroutine, and drains server-initiated messages (elicitation/sampling) to stdout concurrently — the same architecture the HTTP transports already used.
- **`get_test_case_scenario` and `get_test_case`'s `manual_scenario` field always returned an empty step list**, even when `get_test_case_steps` showed a fully populated scenario. Both called `GET /api/testcase/{id}/scenario`, which is marked `deprecated` in the API spec. Switched both to the same normalized-step endpoint `get_test_case_steps` already uses.
- **`update_test_case_step` rejected an `expected_result`-only edit with `400 step.onlyonedetail`** — the API requires `body` to be resent alongside `expected_result` in the same call. The tool now accepts an optional `test_case_id` and looks up the step's current body automatically when only `expected_result` is given.
- **Setting a step's `expected_result` silently didn't take even on a "successful" call** — empirically, the API doesn't create the step's `expectedResultId` on the first PATCH (a second, identical PATCH is needed), and even then the linked child step initially holds a placeholder ("Expected Result") instead of the real text, requiring a further PATCH targeting the child's own step ID. `update_test_case_step` now verifies the outcome (when `test_case_id` is passed) and performs the repair automatically.

### Changed

- **`update_test_case`'s description now warns that writing `manual_scenario` has been observed to silently corrupt step text** (stored as the literal string `"<empty>"` with no error) and recommends `create_test_case_step`/`update_test_case_step` instead, which reliably persist real text.
- Clarified in tool descriptions: `set_test_case_tags`/`bulk_add_test_case_tags` require an existing tag `id` (a name-only entry for a new tag is rejected with 409) — use `create_test_tag` first; `run_test_case` requires the case to already be in the launch via `add_test_cases_to_launch` (otherwise 409); `search_test_cases` AQL string literals must be single-quoted (double quotes return 400 Invalid AQL).

## [2.1.9] - 2026-08-24 - Fix Docker Build on Go 1.27

### Fixed

- **Docker Release build failed** (`go.mod requires go >= 1.27 (running go 1.26.7; GOTOOLCHAIN=local)`) — the builder stage was still pinned to `golang:1.26-alpine` after 2.1.8 bumped `go.mod` to 1.27. Bumped the Dockerfile's base image to `golang:1.27-alpine`. Cross-platform binary builds were unaffected (they don't use the Dockerfile) and shipped correctly in 2.1.8.

## [2.1.8] - 2026-08-24 - Vendored Widgets, Destructive-Op Guard, Test Coverage

### Fixed

- **`execute_testops_operation` could run any DELETE/PUT/PATCH from the 600+ OpenAPI operations with no server-side check.** The `destructiveHint` annotation was advisory metadata only — no client is required to honor it. The tool now requires `parameters.confirm: true` for any operation that resolves to DELETE, PUT, or PATCH; without it, the call fails with a clear error instead of executing.
- **ext-apps widget bundle was fetched from unpkg/jsdelivr at runtime with no integrity check.** A compromised or unavailable CDN could serve unverified JS into the widget sandbox, and a single transient network failure permanently poisoned the process-lifetime cache with a stripped-down fallback stub until restart. The exact pinned `1.7.4` bundle is now vendored into the repo (`internal/tools/assets/ext-apps-1.7.4.js`) and loaded via `go:embed` — no runtime network dependency, no fallback-stub failure mode. A test now fails the build if a future version bump ships a bundle whose export shape the rewrite regex no longer recognizes.
- **Tool count documentation had drifted from the code in multiple ways** (README.md and docs/API.md said 104, llms.txt said both 104 and 114 in the same file) — corrected to 115 everywhere, the real count from a fully-configured server (verified against the actual binary, not just the unit-test registry, which never exercises the OpenAPI-spec-loaded code path since `go test`'s working directory prevents `FindSpecFile` from finding `spec/testops.json`). A new `TestDocsToolCountMatchesRegistry` test now fails the build if the docs and the code disagree again.

### Changed

- **Go version bumped to 1.27** (`go.mod`, CI, and all docs referencing the minimum Go version).
- Widget HTML/JS templates (`launch-dashboard`, `action-picker`, `results-display`) moved from Go string constants into standalone `.html` files under `internal/tools/assets/`, loaded via `go:embed` — gives them real editor syntax highlighting/linting.
- `internal/adapters/allure/client.go` refactored: the repeated build-request → auth → send → status-check → decode pattern across ~110 methods is now a shared `doJSON`/`doRequest`/`doRaw` helper, with behavior unchanged.
- Test coverage raised well above 80% across all internal packages (was as low as 0% in `internal/adapters/allure`, `internal/audit`, `internal/core`, `internal/session`, `internal/tasks`, and ~31% in `internal/mcp`).

## [2.1.7] - 2026-07-14 - Revert SDK Downgrade, Route Around jitless zod Bug Instead

### Fixed

- **`2.1.6`'s downgrade to `ext-apps@1.6.0` broke the widget handshake** with a new error (`i.parts is not iterable`) — that version's wire protocol isn't compatible with the current Claude Desktop host. Reverted the pinned bundle back to `1.7.4`.
- **Internal zod crashes** (`_zod`/`.def` undefined) instead worked around via the SDK's own `allowUnsafeEval: true` App option, which skips the `zod.config({jitless:true})` call in the constructor that appears to trigger them. Applied to all three widgets (dashboard, action picker, results display).

## [2.1.6] - 2026-07-14 - Downgrade Pinned Widget SDK to Avoid jitless zod Crashes

### Fixed

- **Launch Dashboard widget crashing with shifting internal zod errors** (`Cannot read properties of null (reading '_zod')`, `Cannot read properties of undefined (reading 'def')`) on valid, non-null data. Traced to `ext-apps@1.7.x`, which added an unconditional `zod.config({jitless:true})` call in the `App` constructor (needed since the widget iframe's CSP blocks eval-based JIT compilation) — that code path appears to have bugs that surface intermittently depending on which fields are present. Pinned the bundle back to `1.6.0`, which predates that change and has the same `App` API shape our widgets use.

## [2.1.5] - 2026-07-14 - Fix Widget Crash on Null Launch Status

### Fixed

- **Launch Dashboard widget still crashing with `Cannot read properties of null (reading '_zod')` on some launches.** Root cause: Allure returns `status: null` for launches without a terminal status yet (e.g. still running), and that raw `null` was passed straight through `get_launch_dashboard`/`get_launch_details` into the widget's data payload, tripping the host's schema validation. Both tools now normalize status (string, `{id,name}` object, or `null`) into a plain string, defaulting to `"UNKNOWN"`.

## [2.1.4] - 2026-07-14 - Pinned Widget SDK & Better Diagnostics

### Fixed

- **Launch Dashboard widget crashing with `Cannot read properties of null (reading '_zod')`.** The ext-apps SDK bundle was fetched unpinned from unpkg/jsdelivr ("latest"), so an upstream package release could silently change SDK internals underneath us — the same class of issue behind the two prior widget fixes (`2.1.2`, `2.1.3`). The bundle version is now pinned (`1.7.4`) and only bumped deliberately after verification.
- **Widget failures were hard to diagnose.** `new App(...)`/`app.connect()` weren't wrapped in error handling, so an SDK-internal throw could leave the widget blank with no explanation. The whole init flow is now guarded, and both server-side (bundle fetch attempts per CDN candidate) and client-side (`console.error` with the full error/stack) now log the actual reason a widget failed to render instead of just a generic fallback message.

### Fixed

- **Widgets still rendering blank after the ext-apps bundle fix.** The `App` constructor was called with a bare string (`new App('launch-dashboard', {}, {})`) instead of the required `{name, version}` info object, breaking the widget-host connection handshake so the host never revealed the iframe. Also replaced inline `onclick="..."` handlers in the dashboard's action buttons with `addEventListener`, since the widget sandbox's CSP blocks inline event handler attributes. Added `autoResize: true` so widget height tracks rendered content.

## [2.1.2] - 2026-07-14 - Widget Rendering Fix

### Fixed

- **Interactive widgets (dashboard, action picker, results display) rendering blank.** The ext-apps SDK bundle was fetched from `dist/app-with-deps.js`, but the published package places it at `dist/src/app-with-deps.js` — every CDN candidate 404'd, so the server always silently fell back to a minimal stub that doesn't implement the real widget host handshake. Corrected the CDN paths so the real SDK loads.

## [2.1.1] - 2026-07-13 - Safer Deletes & Sturdier Error Handling

### Fixed

- **`remove_test_cases_from_launch` (`mode="delete"`) — no confirmation before permanent deletion.** Now requires interactive elicitation confirmation before deleting test results, matching the existing pattern used by `delete_test_case`/`bulk_delete_test_cases`. Without an interactive session, the call now fails clearly instead of silently deleting, and suggests `mode="hide"`.

- **`add_test_cases_to_launch` — fragile "no-job-assigned" detection.** The friendly error message for automated test cases without a CI job was previously matched via a substring check on the raw upstream error text, which would silently stop firing if Allure changed its wording. Errors from the Allure client are now parsed into a structured `allure.APIError` with a machine-readable `Code` field, checked via `errors.As` (with the substring match kept only as a fallback).

- **`execute_testops_operation` description** — now explicitly tells the model to call `search_testops_operations` first for the exact parameter schema of the target operation, since the generic `parameters` field is necessarily untyped (it proxies 600+ distinct OpenAPI operations).

## [2.1.0] - 2026-06-19 - Project Search, Launch Cleanup & Tool Discoverability

### Added

- **`find_project`** — find a project by name or code (case-insensitive substring match). Resolves a human-readable name/code (e.g. `TSi`) to its numeric project ID without paging through `list_projects` or guessing IDs. Scans `/api/project` client-side since the Allure TestOps API exposes no server-side project name filter; returns `matches`, `count`, `scanned`, and a `truncated` flag when more matches may exist beyond `limit` (default 20) or the 5000-project scan cap.

- **`remove_test_cases_from_launch`** — remove test cases from a launch (e.g. to trim a launch with too many cases or duplicates, including already-started ones). Resolves each `test_case_id` to its test result(s) in the launch — including retries — and removes them. `mode="hide"` (default) excludes the results from the report but keeps the data (`POST /api/testresult/bulk/hide`, non-destructive); `mode="delete"` permanently deletes them (`DELETE /api/testresult/{id}`). Since the Allure API has no bulk test-result delete, `delete` issues one request per result and reports per-result failures under `failed`. Returns `removed_count`, `removed_result_ids`, `not_found_test_case_ids`, and `truncated`.

- **`initialize` now returns an `instructions` field** — the MCP handshake response includes a capability overview (tool groups, how to resolve names→IDs, the `search_testops_operations`/`execute_testops_operation` fallback for the 600+ endpoints, and safety notes). MCP clients add this to the model's system prompt, so Claude knows what the server can do instead of inferring it from bare tool names. Directly improves tool discoverability.

- Tool count: 112 → 114.

### Fixed

- **`llms.txt` listed the removed `update_launch_environment` tool** — dropped from the Launches group (it was removed in 2.0.3) and the group count corrected.

## [2.0.3] - 2026-06-01 - API Compliance Fixes

### Fixed

- **`update_test_case` — `id` missing from request body** — `PATCH /api/testcase/{id}` requires the test case ID both in the URL path and in the JSON body. The body field was absent, causing the API to reject or misroute update requests.

- **`update_test_case` — scenario steps lacked required `type` discriminator** — The Allure API uses a polymorphic step schema (`ScenarioStepDto`) with a required `type` field (`"body"` or `"expected"`). Steps were sent without this field, so the API could not determine the step variant. The `manual_scenario` field is now typed as `ScenarioDto{Steps []ScenarioStepDto}` with `type` enforced. Tool description and schema updated to make the step format explicit.

- **`add_test_case_defect` — wrong endpoint** — was calling `POST /api/testcase/{id}/defect` (GET-only in spec). Corrected to `POST /api/testcase/{testCaseId}/defect/{defectId}` with no request body.

- **`remove_test_case_members` — wrong endpoint and method** — was calling `DELETE /api/testcase/{id}/members` (not in spec). Replaced with `POST /api/testcase/bulk/member/remove` using the correct `{ids, selection}` body. Tool schema updated to require `project_id` (needed for the `selection` object).

- **`resolve_test_result` — missing required `status` field** — `POST /api/testresult/{id}/resolve` requires a `status` body (`failed`, `broken`, `passed`, `skipped`, `unknown`). Was sending no body. Both the client method and tool schema now require `status`.

- **`bulk_resolve_test_results` — missing required `status` field** — `POST /api/testresult/bulk/resolve` requires `status` alongside `selection`. The `TestResultBulkResolveDto` struct and tool schema updated accordingly.

### Removed

- **`update_launch_environment`** — `PUT /api/launch/{id}/env` does not exist in the Allure TestOps API (spec defines only `GET` on this path). Tool and client method removed to prevent silent failures.

## [2.0.2] - 2026-06-01 - Tool Discovery & Test Case Management

### Fixed

- **All tools now visible to Claude** — `tools/list` page size raised from 50 to 1000. With 110+ tools sorted alphabetically, clients that fetch the list once without following `nextCursor` (Claude Desktop, Claude Code) were silently missing every tool past the first 50 — including `update_test_case`, `update_test_case_step`, `validate_test_case_query`, and others starting with letters S–V. All tools are now returned in a single response.

### Added

#### Test Case Tree Navigation
- **`browse_test_case_tree`** — browse test cases at any folder level in the project tree. Pass an empty `path` to start at the root; use returned folder IDs to navigate deeper.
- **`get_test_case_tree_folders`** — list subfolders at a given tree path. Use before moving test cases to find the correct destination.
- **`move_test_cases_to_folder`** — move test cases to a specific folder (`POST /api/testcase/bulk/draganddrop`). The primary tool for organizing and sorting test cases across the project tree.
- **`create_test_case_folder`** — create a new folder (group) at any level of the tree.

#### Scenario Read Tools
- **`get_test_case_scenario`** — read the full step tree of a test case (names, keywords, expected results, nesting). Use before editing steps.
- **`get_test_case_steps`** — read the normalized step list including step IDs required by `move_test_case_step`, `copy_test_case_step`, and `delete_test_case_step`.

#### Prompt
- **`test-case-management`** — workflow prompt for test case CRUD: covers explore → read → edit → organize → safety checkpoints. Accepts `project_id` and optional `goal`.

### Changed

- All tool descriptions rewritten to explain **when** to use each tool, what it returns, and which other tools to call first — rather than just describing the API action.
- `update_test_case` description now explicitly lists editable fields (`name`, `description`, `precondition`, `expected_result`, `status`, `tags`, `members`, `links`, `test layer`) so Claude picks the right tool without guessing.
- `bulk_mute_test_results`, `bulk_unmute_test_results`, `bulk_resolve_test_results`, `bulk_mute_test_cases` descriptions now explain what "mute" and "resolve" mean in TestOps context (excluded from pass rate, alerts suppressed).

## [2.0.1] - 2026-06-01 - Security & Health Check Fixes

### Added
- `GET /health` — unauthenticated liveness probe endpoint; returns `200 ok`. Used by Docker `HEALTHCHECK`, Kubernetes liveness/readiness probes, and uptime monitors. Does not expose any server state.

### Fixed
- **Audit log bypass** — notification-style JSON-RPC requests (no `id` field) were silently skipped by the audit middleware. A client could send `tools/call` without an `id` and the tool would execute without being logged. All requests are now audited regardless of notification status; notification entries get `status: "notification"`.
- **Docker `HEALTHCHECK`** — was calling unauthenticated `GET /sse`, which returns `401` when `MCP_AUTH_TOKENS` is configured. Now calls `GET /health`.
- **CORS wildcard default** — `CORS_ALLOWED_ORIGIN` now defaults to `""` (disabled) instead of `"*"`. Operators must explicitly set the origin. A startup INFO log explains the setting and suggests `https://claude.ai` for browser-based access. Claude Desktop and CLI tools (mcp-remote) are unaffected — they do not enforce browser CORS policy.
- **`/health` method restriction** — previously accepted any HTTP method and returned `200`. Now only `GET` and `HEAD` are accepted; other methods return `405 Method Not Allowed`.

### Changed
- All documentation updated to prefer `MCP_AUTH_TOKENS=name:token,...` over the legacy `MCP_AUTH_TOKEN` single-token variable in shared server examples.
- Kubernetes liveness/readiness probes in docs updated to use `/health` instead of `/sse` or `/messages`.

## [2.0.0] - 2026-05-31 - MCP Protocol Improvements

### Added

#### Generic Typed Handlers
- All 104+ tool handlers now use a `Typed[T]` generic wrapper that auto-deserializes JSON input — eliminates ~200 manual `json.Unmarshal` calls.
- Handler signatures changed from `(ctx, json.RawMessage)` to `(ctx, TypedArgs)` — cleaner and type-safe.

#### slog Integration
- `internal/core.Logger` now wraps `log/slog` with `slog.NewJSONHandler`. Same public API, standard Go logging internals.

#### Middleware Chain + Panic Recovery
- New `internal/mcp/middleware.go` with composable `middlewareFunc` chain.
- **Panic recovery middleware** (first in chain): catches handler panics, logs stack trace, returns `{code: -32603}` instead of crashing the server.
- Audit logging extracted to its own middleware; `dispatch()` simplified to a single call.

#### MCP Protocol 2025-11-25
- `ProtocolVersion` bumped to `"2025-11-25"`.
- Version negotiation in `initialize`: server accepts client's requested version if in supported list (2024-11-05, 2025-03-26, 2025-06-18, 2025-11-25).
- New `logging` and `elicitation` capability fields in `ServerCapabilities`.

#### Cursor-Based Pagination
- `tools/list`, `resources/list`, `prompts/list` now support cursor-based pagination (page size 50).
- Response includes `nextCursor` when more pages exist. Tools list sorted alphabetically for stable cursors.

#### Async Task System
- New `internal/tasks` package: `Task` struct with status lifecycle (working/succeeded/failed/cancelled), in-memory `Store` with context-based cancellation.
- Three new tools: **`get_task_status`**, **`list_running_tasks`**, **`cancel_task`**.
- Long-running operations now return `{task_id, message}` immediately and complete in the background: `run_allure_launch`, `copy_launch`, `merge_launches`, `bulk_run_test_cases_new_launch`, `bulk_run_test_cases_existing_launch`, `bulk_clone_test_cases`.

#### Completion (Argument Autocompletion)
- New `completion/complete` JSON-RPC method.
- Registry has a `Complete(promptName, argName, partial)` method — returns up to 10 suggestions.
- `project_id` arguments complete against live Allure project list when API is configured.

#### Elicitation (Confirmation Dialogs)
- Server can ask the user to confirm destructive operations via `elicitation/create` notification.
- `session.ElicitFunc` context key lets handlers call into the elicitation round-trip.
- Applied to: **`delete_test_case`** and **`bulk_delete_test_cases`** — user must accept before deletion proceeds.
- Route: `notifications/elicitation/complete` delivers user's answer back to the server.

#### Resource Subscriptions
- `resources/subscribe` and `resources/unsubscribe` JSON-RPC methods are now handled.
- `Server.PublishResource(uri)` sends `notifications/resources/updated` to all subscribed sessions.
- Launch dashboard widget auto-starts a 10-second polling watcher when subscribed via `ui://widgets/launch-dashboard?launch_id=N`.
- Server advertises `resources.subscribe = true` in capabilities.

#### Sampling (Server → LLM via Client)
- Server can request LLM inference through the client via `sampling/createMessage`.
- `session.SamplingFunc` context key exposes this to handlers.
- New **`analyze_launch_failures`** tool: fetches failed test results and asks Claude to identify root causes and suggest fixes. Gracefully degrades if sampling is unavailable.

### Fixed

#### Async Task System
- **Panic recovery in goroutines** — `tasks.Store.Run()` wraps every background goroutine in `recover()`. A panic marks the task `Failed` and logs the stack trace; the server and other sessions remain unaffected.
- **Session token lost in async context** — `Store.Create()` now accepts `parentCtx` and uses `context.WithoutCancel` to propagate session ID (and all other context values) without inheriting the request's cancellation signal. Async operations (`run_allure_launch`, `copy_launch`, `merge_launches`, `bulk_run_*`, `bulk_clone_test_cases`) now correctly resolve per-user Allure tokens in multi-user deployments.
- **Data race in task reads** — `Store.Get()` and `Store.List()` now return struct copies under `RLock` instead of raw pointers; `taskToMap` no longer races against concurrent `Update` writes.
- **Async task timeout** — `taskCtx` carries a 30-minute hard deadline via `context.WithTimeout`; hung Allure calls can no longer leak goroutines and memory indefinitely.
- **Task store memory leak** — Added background janitor (`StartJanitor`) that purges succeeded/failed/cancelled tasks older than 1 hour; runs every 5 minutes.

#### MCP Protocol — Elicitation & Sampling
- **Standard JSON-RPC response routing** — `elicitation/create` is now sent as a proper JSON-RPC *request* (with `id`); `sampling/createMessage` already had an `id`. Both now receive the client's standard JSON-RPC response `{id, result}` routed through a new `handleJSONRPCResponse` dispatcher, keyed by the request ID. The previous non-standard methods `notifications/elicitation/complete` and `sampling/createMessage/response` are removed. Compatible with Claude Desktop and any spec-compliant MCP client.
- **Silent deletion on stdio** — `delete_test_case` and `bulk_delete_test_cases` on the stdio transport previously bypassed the confirmation guard (no `ElicitFunc` in context → the `if ok` block was skipped, and data was deleted silently). Now they return an error: *"deletion requires user confirmation but no interactive session is available"*. Also, transport errors from `elicit()` are now returned as errors instead of being masked as a user cancellation.

#### Resource Subscriptions
- **Push notifications never arrived** — `handleResourcesSubscribe` was passing the HTTP request context to `OnSubscribe → StartLaunchWatch → watchLaunch`. The request's `r.Context()` was cancelled when the `POST /messages` handler returned (202 Accepted), so the watcher exited before the first 10-second tick. Now `sess.ctx` (session lifetime) is used — the watcher runs until the client disconnects.

#### Other
- **Unstable pagination** — `resources/list` and `prompts/list` now sort by URI and name respectively before slicing, matching `tools/list` behaviour. Previously, cursor-based pagination over Go maps produced non-deterministic results.
- **Completion hits Allure on every keystroke** — `Complete` now maintains a 30-second in-process cache of project/launch ID lists fetched with a 5-second timeout derived from the request context. Completion is gated on `ref.type == "ref/prompt"` to avoid unintended Allure calls from resources with identically-named arguments.
- **UTF-8 truncation in analysis** — `analyzeLaunchFailures` truncated error messages and stack traces by byte index (`msg[:300]`), which split multi-byte runes (Cyrillic, CJK, emoji). Now uses `truncateRunes` which truncates at rune boundary.

### Changed
- Tool count: 100 → 104 (added `get_task_status`, `list_running_tasks`, `cancel_task`, `analyze_launch_failures`).
- `initialize` response now includes `logging` capability.
- `resources.subscribe` is now `true` in `initialize` capabilities.

## [1.9.0] - 2026-05-31 - MCP Prompts & Resources

### Added

#### MCP Prompts
- **`analyze-test-failures`** — prompt template that instructs Claude to retrieve launch details, list failed/broken results, group by error pattern, and surface critical issues. Argument: `launch_id` (required), `project_id` (optional).
- **`launch-report-summary`** — prompt template that instructs Claude to generate a concise executive summary (pass rate, counts, duration, environment, top failures). Argument: `launch_id` (required), `project_id` (optional).
- Server now advertises `prompts` capability in the MCP `initialize` response.
- `prompts/list` and `prompts/get` JSON-RPC methods are fully handled.

#### MCP Resources
- **`allure://docs/quickstart`** — static Markdown resource listing tool groups, prompt templates, and quick-start steps. Available even when no Allure token is configured; clients can attach it as context.

## [1.8.1] - 2026-05-30 - Doc Fixes

### Fixed
- **Claude Desktop client config** — replaced `url` + `headers` format (not supported by Claude Desktop for custom headers) with `mcp-remote` proxy pattern. All documentation examples now show `npx mcp-remote <url> --header Key:Value` for both Windows (`cmd /c npx ...`) and macOS/Linux (`npx ...`). Updated README, DEPLOYMENT.md, QUICKSTART.md, llms.txt, llms-full.txt.

## [1.8.0] - 2026-05-30 - Multi-User Auth & Audit Log

### Added

#### Multi-User Authentication
- **`MCP_AUTH_TOKENS`** environment variable — configure named user tokens in `name:token,...` format (e.g. `alice:abc123,bob:xyz789`). Each user authenticates with their own bearer token; requests are attributed to the user by name in logs and audit records.
- **Backward-compatible** — existing `MCP_AUTH_TOKEN` (single token) still works and is treated as user `"default"`.
- Startup log now shows the configured user count instead of a boolean auth flag.

#### Audit Log
- **Daily JSONL audit files** written to a configurable directory (`AUDIT_LOG_PATH`, default `audit/`).
- Each entry records: `timestamp`, `user`, `session_id`, `remote_addr`, `method`, `tool` (for `tools/call`), `status` (`ok`/`error`), `duration_ms`.
- **Automatic retention** — files older than `AUDIT_RETENTION_DAYS` days (default 30) are deleted nightly.
- Docker Compose mounts `./audit` as a host volume so logs survive container restarts.
- Audit logger is disabled gracefully (warning logged) if the directory cannot be created.

#### New Environment Variables
| Variable | Default | Description |
|---|---|---|
| `MCP_AUTH_TOKENS` | — | Named user tokens: `alice:tok1,bob:tok2` |
| `AUDIT_LOG_PATH` | `audit` | Directory for daily audit JSONL files |
| `AUDIT_RETENTION_DAYS` | `30` | Days to keep audit files |

### Fixed

#### GitHub Actions — Docker Release
- Added **QEMU** setup step (`docker/setup-qemu-action@v3`) required for multi-platform (`linux/amd64`, `linux/arm64`) builds — previously the multi-arch build was silently skipped.
- Added `platforms: linux/amd64,linux/arm64` to `build-push-action`.
- Added `type=raw,value=latest` tag so the `:latest` image is correctly pushed on tag releases.
- Fixed `publish-release` step using `--notes-append` instead of `--notes-file` (which was overwriting the full release body).
- Pinned action versions to stable releases (`actions/checkout@v4`, `docker/setup-buildx-action@v3`, `docker/login-action@v3`, `docker/metadata-action@v5`).

### Changed
- `docker-compose.yml` — removed duplicated `environment` keys that were already covered by `env_file`; added `AUDIT_LOG_PATH` and `AUDIT_RETENTION_DAYS` with defaults; added `./audit:/app/audit` volume mount.

## [1.7.1] - 2026-05-29 - MCP Compliance Fixes & Model Expansions

### Fixed

#### MCP Protocol Compliance
- **Tool annotations missing on 3 tools** — `configure_allure_token`, `search_testops_operations`, and `execute_testops_operation` were registered after the annotation loop and sent `null` annotations to Claude. All three now receive correct `readOnlyHint`/`destructiveHint` values.
- **`execute_testops_operation` now marked `destructiveHint: true`** — the tool can execute any API operation including DELETE/PUT, which requires the destructive hint per MCP spec.
- **CORS headers missing custom headers** — `Mcp-Session-Id` and `X-Allure-Token` added to `Access-Control-Allow-Headers`; `DELETE` added to `Access-Control-Allow-Methods`. Previously, browsers would block cross-origin preflight requests using these headers.
- **Streamable HTTP Content-Type validation** — POST `/mcp` now rejects requests with a non-JSON `Content-Type` with `415 Unsupported Media Type`, as required by MCP spec 2025-03-26.
- **`GetLaunchStatistics` response parsing** — the Allure API returns a `[]{"status", "count"}` array, not a single object. Client now aggregates the array into `StatisticsResponse` correctly; the Launch Dashboard widget now shows real pass/fail/broken counts.

#### Data Models
- **`ClientCapabilities`** — `InitializeRequest.Capabilities` was an empty `struct{}`, silently discarding client capability flags. Replaced with typed struct parsing `elicitation`, `sampling`, and `roots` fields — required for future elicitation support.
- **`LaunchCreateRequest`** expanded with `AutoClose`, `External`, `Issues`, `Links`, `Tags` fields matching the full API schema.
- **`LaunchResponse`** enriched with `UUID`, `CreatedDate`, `LastModifiedDate`, `AutoClose`, `Closed`, `External`, `Issues`, `Links`, `Tags`.
- **`StatisticsResponse`** expanded with `Unknown` field; new `StatisticItem` DTO for the raw array items from the statistics endpoint.

### Added

#### New Model DTOs
- `CategoryDto` — test result category reference
- `TestLayerDto` — test layer reference  
- `JobRunDto` — CI/CD job run linked to a test result (with URL)
- `IdAndNameOnlyDto` — lightweight id+name reference
- `IntegrationTypeDto` — issue tracker integration type
- `RoleDto` — user role in a project
- `StatusDto` — named status object for test cases
- `WorkflowRowDto` — workflow attached to a test case
- `CustomFieldValueWithCfDto` — payload for setting custom field values via the `/cfv` endpoint

### Notes
- No tool count changes — all fixes are internal correctness and protocol compliance improvements
- Fully backwards-compatible: existing Claude Desktop and claude.ai configurations require no changes

## [1.7.0] - 2026-05-27 - Interactive API Discovery Widgets

### Added

#### MCP App Widgets for Search/Execute Tools
- **Action Picker Widget** — interactive searchable picker for `search_testops_operations` results
  - Real-time filtering by operation name, description, or API path
  - HTTP method color-coding (GET green, POST blue, PUT orange, DELETE red)
  - Click to select operation and inject operation_id into chat
- **Results Display Widget** — formatted result viewer for `execute_testops_operation` 
  - Status indicator (Success/Error) with color
  - Auto-formatted JSON with proper indentation
  - Scrollable body for large API responses
  - Graceful fallback for raw text responses

### Changed
- **`search_testops_operations`** — updated to render Action Picker widget; description now mentions widget support
- **`execute_testops_operation`** — updated to render Results Display widget; description now mentions widget support

### Notes
- All widgets support light/dark mode via host theme detection
- Compatible with Claude Desktop and claude.ai
- Uses ext-apps bundle for cross-platform rendering

## [1.6.0] - 2026-05-26 - Full Test Case API Coverage

### Added

#### Test case single-item operations (25 new tools)
- **`get_test_case_tags`** / **`set_test_case_tags`** — read and replace tags on a test case
- **`get_test_case_issues`** / **`set_test_case_issues`** — read and replace linked bug-tracker issues
- **`get_test_case_examples`** / **`set_test_case_examples`** — parametrized data table rows (full replace)
- **`list_test_case_versions`** / **`create_test_case_version`** / **`restore_test_case_version`** — version snapshot management
- **`get_test_case_version_data`** / **`delete_test_case_version`** — version content and cleanup
- **`get_test_case_attachments`** / **`delete_test_case_attachment`** — attachment listing and removal
- **`search_test_cases`** — AQL/RQL full-text search across project test cases
- **`list_deleted_test_cases`** — browse soft-deleted test cases
- **`list_muted_test_cases`** — browse muted test cases
- **`delete_test_case_scenario`** — remove the entire step scenario from a test case
- **`move_test_case_step`** / **`copy_test_case_step`** — reposition steps within a scenario
- **`get_test_case_relations`** / **`set_test_case_relations`** — test-case-to-test-case relations
- **`get_test_case_custom_fields`** / **`update_test_case_custom_fields`** — custom field values via dedicated `/cfv` endpoint
- **`get_test_case_workflow`** — workflow definition for a test case
- **`get_test_case_keys`** / **`set_test_case_keys`** — integration test keys (Jira, Azure DevOps, etc.)
- **`get_test_case_scenario_from_run`** — scenario captured from the last automated run
- **`detach_test_case_automation`** — convert an automated test case back to manual
- **`get_test_case_audit`** — change history / audit log
- **`validate_test_case_query`** — validate AQL/RQL expression without executing it
- **`suggest_test_cases`** — autocomplete suggestions by name

#### Bulk test-case operations (14 new tools)
- **`bulk_add_test_case_members`** / **`bulk_remove_test_case_members`**
- **`bulk_add_test_case_custom_fields`** / **`bulk_remove_test_case_custom_fields`**
- **`bulk_add_test_case_external_links`**
- **`bulk_add_test_case_issues`** / **`bulk_remove_test_case_issues`**
- **`bulk_set_test_case_layer`**
- **`bulk_move_test_cases`** — move test cases to another project
- **`bulk_delete_test_cases`** — permanent bulk delete
- **`bulk_run_test_cases_new_launch`** / **`bulk_run_test_cases_existing_launch`**
- **`bulk_create_test_plan`** — create a test plan from selected test cases
- **`bulk_mute_test_cases`**

### Fixed
- **`execute_testops_operation`** parameter routing — path parameters no longer leaked into the request body; unknown parameters are now collected separately
- **`execute_testops_operation` array bodies** — new explicit `body` key lets callers pass any JSON value (array or object) directly as the HTTP request body

### Changed
- Total MCP tools: **57 → 102** (45 new tools, all backed by proper client methods — no OpenAPI dynamic execution fallback)
- All new client methods added to `internal/adapters/allure/client.go` with typed request/response models in `models.go`
- New file `internal/tools/tools_testcases_extra.go` for the extended single-item test case tools

## [1.5.0] - Security & Architecture Hardening

### Added
- **Per-session Allure token isolation** — Each SSE session stores its own token; concurrent users on a shared server never mix credentials
- **`X-Allure-Token` header support** — Users pass their personal Allure token from Claude Desktop config via HTTP header; no need to call `configure_allure_token` manually
- **Token-keyed JWT cache** — JWT tokens cached per API key, not globally; separate users get separate JWTs
- **`internal/session` package** — Shared context helpers for session ID propagation across packages
- **Server startup warning** — Logged when HTTP mode runs without `MCP_AUTH_TOKEN`
- **Build-time version injection** — Server version set via `-ldflags` instead of hardcoded `"1.0.0"`

### Changed
- **`registry.go` split into domain files** — `tools_launches.go`, `tools_results.go`, `tools_testcases.go`, `tools_projects.go`, `tools_analytics.go`, `tools_bulk.go`, `tools_relations.go`
- **`GetSessionToken` callback now context-aware** — Receives `context.Context` to resolve the correct per-session token
- **Dockerfile** — Pinned `alpine:3.21`, removed `.exe` suffix from Linux binary
- **HTTP server timeouts** — Added `ReadTimeout` and `IdleTimeout`; `WriteTimeout` disabled for SSE long-lived streams
- **Documentation** — All Claude Desktop config examples corrected to use `url` + `headers` for remote servers; `Authorization` marked optional

### Fixed
- **Data race on JWT cache** — Added `sync.Mutex` protecting `jwtToken`/`jwtExpiresAt` fields
- **Cross-user token contamination** — Global `sessionToken` field replaced with per-session map
- **Invalid JSON on marshal error** — `resultToJSON` error path now uses `json.Marshal` instead of `fmt.Sprintf`
- **Predictable session ID fallback** — `crypto/rand` failure now panics instead of using `time.Now().UnixNano()`
- **`SetSessionTokenFunc` visibility** — Exported function field replaced with proper setter method

## [1.4.0] - Token Priority & Shared Server Auth

### Added
- **Improved token handling for shared servers** — Users can now provide personal tokens on shared server deployments
- **Session token priority** — Personal tokens override server tokens, enabling per-user authentication on shared setups

### Changed
- **ALLURE_TOKEN handling** — Server token is now truly optional (session token takes priority)
- **README reorganized** — Cleaner quick start for end users, separated setup guides for per-user and shared server setups
- **.env.example updated** — Clear documentation of per-user vs shared server setup modes with helpful comments

### Fixed
- Token priority logic in `client.go` — session/user tokens now correctly override fallback server tokens
- Documentation clarity on shared server setup without mandatory ALLURE_TOKEN

## [1.3.0] - 2026-05-12

### Added
- **Search + Execute tools** — Query and execute any of 600+ TestOps API endpoints dynamically
- **Per-user authentication** — Users can authenticate with their own tokens in chat, no shared credentials needed
- **Session-based token configuration** — New `configure_allure_token` tool for chat-based auth setup with security warnings
- **Go 1.26 support** — Updated all tooling, CI/CD, and Docker builds to Go 1.26

### Changed
- **ALLURE_TOKEN is now optional** — Shared server deployments no longer require pre-configured tokens
- **OpenAPI spec bundled** — Complete Allure TestOps API specification included in repo (spec/testops.json)
- **Improved documentation** — Added per-user vs shared server setup guides, lazy setup instructions

### Improved
- Authentication flexibility for different deployment scenarios
- API coverage expanded from 55 explicit tools to 55 + dynamic search/execute for 600+ operations
- Development experience with updated Go dependencies across the board

## [1.2.1] - 2026-05-08

### Added
- **GitHub Actions CI/CD pipeline** — Automated multi-platform builds for Windows, macOS, and Linux
- **CLAUDE.md** — Developer documentation with release procedures and project guidelines

### Changed
- **Simplified README** — Local Claude Desktop setup now the primary onboarding flow
- **Improved getting-started experience** — 3-minute setup instead of 10+ minutes
- **Marketing-focused documentation** — Value-driven introduction, use cases, and comparisons
- **Better documentation structure** — Clearer navigation and links throughout

### Improved
- README with benefit-driven messaging
- Deployment option clarity with examples
- Contributing guidelines
- SEO optimization for discoverability
- Build artifact management (.gitignore updates)

### Infrastructure
- Automated release process: `git tag vX.Y.Z && git push origin vX.Y.Z`
- Pre-built binaries attached automatically to GitHub releases
- CI/CD pipeline ready for future releases

## [1.2.0] - 2026-05-08

### Added
- **21 new MCP tools** for comprehensive test case management
- Complete defect management: `add_test_case_defect`, `remove_test_case_defect`, `get_test_case_defects`, `get_launch_defects`
- Team collaboration: `get_test_case_members`, `add_test_case_members`, `remove_test_case_members`
- External links: `get_test_case_external_links`, `add_test_case_external_link`, `delete_test_case_external_link`
- Enhanced test case operations: `clone_test_case`, `restore_test_case`, `get_test_case_history`
- Advanced launch operations: `copy_launch`, `merge_launches`
- Environment management: `get_launch_environment`, `update_launch_environment`
- Bulk operations: `bulk_clone_test_cases`
- Single test result operations: `resolve_test_result`, `unmute_test_result`
- Manual scenario update support via `update_test_case` tool

### Changed
- Tool count: 35 → 55 (added 20 new tools)
- Release distribution: pre-built binaries for Windows, macOS (Intel & ARM), and Linux
- Updated README with quick start for pre-built binaries

### Added Documentation
- New `RELEASES.md` file with detailed setup instructions for all platforms
- Platform-specific guides: Windows, macOS, Linux
- Troubleshooting section for common issues
- API token retrieval instructions

### Features
- ✅ Support for `manual_scenario` update in test case updates
- ✅ Comprehensive test case lifecycle management
- ✅ Full defect tracking integration
- ✅ Team member assignment and management
- ✅ External link/relation support (GitHub, Jira, etc.)
- ✅ Launch environment variable management
- ✅ Test case cloning and restoration
- ✅ Launch merging capabilities

## [1.1.2] - 2026-05-08

### Fixed
- `get_test_case` now includes `manual_scenario` field with complete test execution steps
- Fetches scenario via dedicated `/api/testcase/{id}/scenario` endpoint
- Returns scenario attachments, step structure, and expected results

### Changed
- Tool count: 32 → 35 (added 3 step management tools)
- Tool description clarified to explicitly mention `manual_scenario` field
- README updated with note about accessing test execution steps

### Documentation
- Added explicit note that `get_test_case` includes all test execution steps in `manual_scenario` field
- Updated tools table to include step management tools (`create_test_case_step`, `update_test_case_step`, `delete_test_case_step`)

## [1.1.1] - 2026-05-08

### Fixed
- `get_test_case` now returns complete test case overview from `/api/testcase/{id}/overview` endpoint
- Added execution steps (scenario), tags, members, custom fields, and all metadata to test case details
- Resolved missing Execution section that was not available in basic endpoint

## [1.1.0] - 2026-05-08

### Added
- **Full test case field editing** — Extended `update_test_case` tool to support all fields from TestCasePatchV2Dto:
  - Text fields: `description`, `precondition`, `expected_result`, `full_name`
  - Boolean flags: `automated`, `external`, `deleted`
  - Resource IDs: `status_id`, `test_layer_id`, `workflow_id`
  - Collections: `tags` (array of tag objects), `members` (array of member objects), `links` (array of external links)
- **Step management tools** — Full CRUD operations for test case steps:
  - `create_test_case_step` — Create a new step with optional positioning (`after_id`) and nesting (`parent_id`)
  - `update_test_case_step` — Update step body and expected results
  - `delete_test_case_step` — Delete a step from a test case
- **Extended models** — New DTOs for step operations and external links support

### Changed
- `update_test_case` handler refactored to accept struct-based request instead of individual parameters
- Tool count increased from 32 to 35

## [1.0.0] - 2026-04-23

### Added
- **32 MCP tools** covering the full Allure TestOps workflow:
  - Launch management: `run_allure_launch`, `get_launch_status`, `get_launch_report`, `list_launches`, `get_launch_details`, `close_launch`, `reopen_launch`, `add_test_cases_to_launch`, `add_test_plan_to_launch`
  - Test results: `list_test_results`, `get_test_result`, `assign_test_result`, `mute_test_result`, `bulk_assign_test_results`, `bulk_mute_test_results`, `bulk_unmute_test_results`, `bulk_resolve_test_results`
  - Test cases: `list_test_cases`, `get_test_case`, `create_test_case`, `update_test_case`, `delete_test_case`, `run_test_case`, `bulk_set_test_case_status`, `bulk_add_test_case_tags`, `bulk_remove_test_case_tags`
  - Projects & analytics: `list_projects`, `get_project`, `get_project_stats`, `get_launch_trend_analytics`, `get_launch_duration_analytics`, `get_test_success_rate`
- **Dual transport modes**: stdio (Claude Desktop) and HTTP/SSE (team deployment)
- **Bearer-token authentication** via `MCP_AUTH_TOKEN` for HTTP mode
- **CORS support** with configurable `CORS_ALLOWED_ORIGIN`
- **Structured JSON logging** to stderr with configurable level (`LOG_LEVEL`)
- **Multi-stage Dockerfile** with non-root user and health check
- **Docker Compose** configuration for team deployment with resource limits
- **Kubernetes manifest** (`k8s-manifest.yaml`) with deployment, service, and resource constraints
- **Caddy reverse proxy** config for automatic HTTPS
- **Systemd service** example for Linux deployments
- **MCP protocol 2024-11-05** with full JSON-RPC 2.0 compliance
- Comprehensive documentation: Installation, Deployment, API Reference, Security guides
- `.env.example` configuration template

[Unreleased]: https://github.com/MimoJanra/TestOpsMCP/compare/v2.0.3...HEAD
[2.0.3]: https://github.com/MimoJanra/TestOpsMCP/compare/v2.0.2...v2.0.3
[2.0.2]: https://github.com/MimoJanra/TestOpsMCP/compare/v2.0.1...v2.0.2
[2.0.1]: https://github.com/MimoJanra/TestOpsMCP/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.9.0...v2.0.0
[1.9.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.8.1...v1.9.0
[1.8.1]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.8.0...v1.8.1
[1.8.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.7.1...v1.8.0
[1.7.1]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.7.0...v1.7.1
[1.7.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.6.0...v1.7.0
[1.6.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.5.0...v1.6.0
[1.5.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.2.1...v1.3.0
[1.2.1]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.1.2...v1.2.0
[1.1.2]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/MimoJanra/TestOpsMCP/releases/tag/v1.0.0
