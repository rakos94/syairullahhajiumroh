#!/usr/bin/env bash
# Backup / restore for MongoDB + uploaded documents, used by the `backup` and
# `restore` services in docker-compose.yml.
#
#   backup.sh schedule        run a backup every day at $BACKUP_TIME (HH:MM)
#   backup.sh now             run one backup immediately
#   backup.sh list            list existing backups
#   backup.sh restore <name>  restore a backup (DROPS current data)
#
# Each backup is a folder backups/backup-YYYYmmdd-HHMMSS/ with
# mongo.archive.gz (mongodump) and uploads.tar.gz. Only the newest
# $BACKUP_KEEP backups are kept; older ones are deleted (rotation).
set -euo pipefail

BACKUP_ROOT=/backups
UPLOADS=/uploads
KEEP="${BACKUP_KEEP:-7}"
BACKUP_TIME="${BACKUP_TIME:-02:00}"
: "${MONGO_URI:?MONGO_URI is required}"
: "${MONGO_DB:?MONGO_DB is required}"

log() { echo "[backup] $(date '+%Y-%m-%d %H:%M:%S') $*"; }

rotate() {
  local all n
  mapfile -t all < <(find "$BACKUP_ROOT" -maxdepth 1 -type d -name 'backup-*' | sort)
  n=${#all[@]}
  if (( n > KEEP )); then
    for d in "${all[@]:0:n-KEEP}"; do
      rm -rf -- "$d"
      log "rotasi: hapus backup lama $(basename "$d")"
    done
  fi
}

backup_now() {
  local ts tmp final
  ts=$(date +%Y%m%d-%H%M%S)
  tmp="$BACKUP_ROOT/.tmp-$ts"
  final="$BACKUP_ROOT/backup-$ts"
  # Remove leftovers of interrupted runs.
  find "$BACKUP_ROOT" -maxdepth 1 -type d -name '.tmp-*' -exec rm -rf {} +
  mkdir -p "$tmp"
  mongodump --uri="$MONGO_URI" --db="$MONGO_DB" --archive="$tmp/mongo.archive.gz" --gzip --quiet
  tar -czf "$tmp/uploads.tar.gz" -C "$UPLOADS" .
  # Two backups in the same second must not end up nested in one folder.
  local n=1
  while [[ -e "$final" ]]; do n=$((n + 1)); final="$BACKUP_ROOT/backup-$ts-$n"; done
  mv "$tmp" "$final"   # only complete backups get the backup-* name
  log "berhasil: $(basename "$final") ($(du -sh "$final" | cut -f1))"
  rotate
}

schedule() {
  if ! [[ "$BACKUP_TIME" =~ ^([01][0-9]|2[0-3]):[0-5][0-9]$ ]]; then
    log "BACKUP_TIME tidak valid: $BACKUP_TIME (format HH:MM)"; exit 1
  fi
  log "jadwal: setiap hari pukul $BACKUP_TIME ($TZ), simpan $KEEP backup terakhir"
  while true; do
    local now target
    now=$(date +%s)
    target=$(date -d "today $BACKUP_TIME" +%s)
    (( target <= now )) && target=$(date -d "tomorrow $BACKUP_TIME" +%s)
    log "backup berikutnya: $(date -d "@$target" '+%Y-%m-%d %H:%M')"
    sleep $(( target - now ))
    backup_now || log "GAGAL: backup tidak selesai, lihat error di atas"
  done
}

list() {
  local found=0
  for d in $(find "$BACKUP_ROOT" -maxdepth 1 -type d -name 'backup-*' | sort); do
    echo "$(basename "$d")  $(du -sh "$d" | cut -f1)"
    found=1
  done
  (( found )) || echo "(belum ada backup)"
}

restore() {
  local name="${1:-}" dir
  [[ -n "$name" ]] || { echo "pakai: backup.sh restore <nama-backup>"; list; exit 1; }
  dir="$BACKUP_ROOT/$(basename "$name")"
  [[ -f "$dir/mongo.archive.gz" && -f "$dir/uploads.tar.gz" ]] || { log "backup tidak ditemukan: $name"; exit 1; }
  log "restore database dari $(basename "$dir") (data sekarang DIGANTI)"
  mongorestore --uri="$MONGO_URI" --archive="$dir/mongo.archive.gz" --gzip --drop --quiet \
    --nsInclude="$MONGO_DB.*"
  log "restore uploads"
  find "$UPLOADS" -mindepth 1 -delete
  tar -xzf "$dir/uploads.tar.gz" -C "$UPLOADS"
  log "restore selesai"
}

mkdir -p "$BACKUP_ROOT"
case "${1:-}" in
  schedule) schedule ;;
  now) backup_now ;;
  list) list ;;
  restore) restore "${2:-}" ;;
  *) echo "pakai: backup.sh {schedule|now|list|restore <nama>}"; exit 1 ;;
esac
