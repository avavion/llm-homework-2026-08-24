# BE-002 Regulatory Rules Registry Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish an evidence-backed registry of date rules and product alert groups, disabling automation when primary evidence is missing.

**Architecture:** Two version-controlled documents form the source of truth: one maps ISO countries and regulator groups to date semantics, the other defines initial alert groups and their rationale. An unverified row is explicitly marked `research_required` rather than inferred.

**Tech Stack:** Markdown, primary legal and regulator sources.

**Spec:** `backend/docs/tasks/BE-002-regulatory-rules-registry.md`

## Global Constraints

- Include CIS and EU countries only as documented by their regulator-group evidence.
- Do not invent an EAEU 00:00 expiry rule.
- The minimum user-selectable e-mail threshold is 60 minutes; the policy is not food-safety advice.

---

### Task 1: Create the date-rule evidence registry

**Files:**
- Create: `shared/docs/regulatory-date-rules.md`
- Test: row-completeness review

**Interfaces:**
- Produces: fields `regulator_group`, ISO country codes, `date_type`, `expiry_timezone_source`, `expiry_instant_rule`, `post_expiry_status`, `recipe_eligibility`, source URL, access date, and research status.

- [ ] **Step 1: Gather primary evidence**

For each regulator group, record the direct source and the access date. Use the EU Regulation 1169/2011 and European Commission date-marking guidance already cited by `shared/docs/product-description.md` for the EU distinction.

- [ ] **Step 2: Write the registry table**

Include an enabled row only when every interface field is supported by a primary source. Put incomplete groups in rows with `research_required`, no automatic status, and no schedule.

- [ ] **Step 3: Review evidence gates**

Run: `rg -n "regulator_group|research_required|expiry_instant_rule|recipe_eligibility|https?://" shared/docs/regulatory-date-rules.md`

Expected: every enabled row has a source URL and access date; every uncertain row disables automation.

### Task 2: Define product-group alert policy

**Files:**
- Create: `shared/docs/product-group-alert-policy.md`
- Test: policy-completeness review

**Interfaces:**
- Produces: initial groups `refrigerated_perishable`, `fresh_produce`, `frozen`, `shelf_stable`, and `other`, each with a default alert window, 60-minute minimum, and rationale.

- [ ] **Step 1: Write the policy table**

For each of the five groups, state the default window and why it is a convenience default rather than safety advice. State that users cannot set a threshold below 60 minutes.

- [ ] **Step 2: Check for unsupported automation**

Run: `rg -n "60 минут|research_required|автомат" shared/docs/regulatory-date-rules.md shared/docs/product-group-alert-policy.md`

Expected: the 60-minute limit and research gate are explicit in both documents where applicable.

- [ ] **Step 3: Commit**

Run: `git add shared/docs/regulatory-date-rules.md shared/docs/product-group-alert-policy.md && git commit -m "docs: add regulatory date-rule registry"`

## Self-review

- Coverage: source, access date, country code, date semantics, expiry moment, recipe eligibility, and alert policy are captured.
- No unverified legal or medical conclusion is enabled.

