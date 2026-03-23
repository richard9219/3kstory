#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
API_BASE="${API_BASE:-http://localhost:8080/api/v1}"
PROJECT_ID="${PROJECT_ID:-}"
TOKEN="${TOKEN:-}"
SCENE_ID="${SCENE_ID:-}"
STORYBOARD_FILE="${STORYBOARD_FILE:-$ROOT_DIR/docs/17-药命效应-storyboard-v1.json}"
REQUESTS_FILE="${REQUESTS_FILE:-$ROOT_DIR/docs/18-药命效应-comfy首批6镜头请求.json}"

if [[ -z "$PROJECT_ID" ]]; then
  echo "PROJECT_ID is required"
  exit 1
fi

if [[ -z "$TOKEN" ]]; then
  echo "TOKEN is required"
  exit 1
fi

if [[ ! -f "$STORYBOARD_FILE" ]]; then
  echo "Storyboard file not found: $STORYBOARD_FILE"
  exit 1
fi

if [[ ! -f "$REQUESTS_FILE" ]]; then
  echo "Requests file not found: $REQUESTS_FILE"
  exit 1
fi

AUTH_HEADER="Authorization: Bearer $TOKEN"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "==> Importing storyboard shots into project $PROJECT_ID"
curl --fail --silent --show-error \
  -X POST "$API_BASE/projects/$PROJECT_ID/storyboard-shots/import" \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  --data @"$STORYBOARD_FILE"
echo

if [[ -z "$SCENE_ID" ]]; then
  echo "==> Resolving fallback scene_id from project scenes"
  SCENE_ID="$(curl --fail --silent --show-error \
    -X GET "$API_BASE/projects/$PROJECT_ID/scenes" \
    -H "$AUTH_HEADER" | node -e '
      const fs = require("fs");
      const raw = fs.readFileSync(0, "utf8");
      const payload = JSON.parse(raw);
      const scenes = payload.data || payload.scenes || [];
      process.stdout.write(scenes[0] ? String(scenes[0].id) : "");
    ')"
fi

if [[ -z "$SCENE_ID" ]]; then
  echo "No scene_id resolved. Set SCENE_ID manually or create scenes first."
  exit 1
fi

echo "==> Using scene_id=$SCENE_ID for validation requests"

node - "$REQUESTS_FILE" "$SCENE_ID" "$TMP_DIR" <<'NODE'
const fs = require('fs');
const path = require('path');

const [requestsPath, sceneId, outDir] = process.argv.slice(2);
const payload = JSON.parse(fs.readFileSync(requestsPath, 'utf8'));
const shared = payload.shared_extra_data || {};
const workflowPath = payload.workflow_path;

for (const req of payload.requests || []) {
  const body = {
    scene_id: Number(sceneId),
    provider: payload.provider || 'comfy',
    workflow_path: workflowPath,
    prompt: req.prompt,
    image_url: req.image_url,
    duration: req.duration,
    aspect_ratio: req.aspect_ratio || payload.aspect_ratio || '16:9',
    resolution: req.resolution || '720p',
    extra_data: shared,
  };
  const file = path.join(outDir, `${String(req.shot_number).padStart(2, '0')}-${req.title}.json`);
  fs.writeFileSync(file, JSON.stringify(body, null, 2));
}
NODE

echo "==> Submitting 6 validation shots"
for request_file in "$TMP_DIR"/*.json; do
  echo "--> $(basename "$request_file")"
  curl --fail --silent --show-error \
    -X POST "$API_BASE/projects/$PROJECT_ID/generate-video" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    --data @"$request_file"
  echo
done

echo "==> Done"
