# Backend API gaps discovered while completing the frontend

The frontend uses the contracts from `api/v1/*.yaml`. Blocks that cannot be backed by the current API are explicitly marked **MVP / mock** in the UI.

1. **Current user profile** — add `GET /users/{user_id}/profile` (and/or `/users/me/profile`) returning all competency clusters and indicators. `GET /users/{user_id}/profile/history` only provides a range comparison and is not a canonical current-profile endpoint.
2. **Training history/list** — add `GET /users/{user_id}/training-sessions` (with status/date/mode/scenario/version) so learner and curator can browse completed/active sessions and then open `/training/sessions/{session_id}` + `/analysis`.
3. **Training assignments read API** — add `GET /users/{user_id}/assignments` and preferably curator/team filtering. The backend currently has only `POST /assignments`, so learner cannot retrieve curator-assigned trainings and curator cannot verify assignment status/history.
4. **Recommendation** — add `GET /users/{user_id}/recommendation` (or `/users/me/recommendation`) if the product wants the next recommended training to come from backend rather than a UI placeholder.
5. **Curator scenario permissions** — the current Scenario write/generate/validate/publish/archive endpoints require `active role=methodist`. The requested curator-side constructor needs those permissions expanded to curator or a separate curator-authoring permission/role.
6. **`partner_behavior` in Scenario API** — `AGENT.md` describes this structure for `virtual_partner`, but `scenarios.yaml` does not contain it. Add the field/schema so the constructor can persist AI-partner persona/policies. The frontend currently marks this editor as local-only MVP data.
7. **Curator live spectator WebSocket** — REST session/messages access is sufficient for polling, but a documented read-only curator/spectator WS channel would provide true real-time monitoring without polling. The frontend currently refreshes Arena messages every 2.5 seconds.
8. **Voice/audio live monitoring** — current domain says chat-only MVP. To support curator listening live and replaying audio, add voice session/media/recording contracts and authorization. The current UI explicitly shows this as an MVP placeholder unless `Arena history details.media` already returns media.

Optional optimization: an aggregate `GET /users/{user_id}/arena-feedback` would avoid N requests over Arena history when rendering “all feedback”, although the current API can already assemble this from per-session history details.
