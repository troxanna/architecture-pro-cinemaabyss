#!/bin/sh
set -e

# значения по умолчанию
: "${GRADUAL_MIGRATION:=false}"
: "${MOVIES_MIGRATION_PERCENT:=0}"
: "${EVENTS_MIGRATION_PERCENT:=0}"

# ограничим проценты 0..100 (на всякий)
clamp() {
  v="$1"
  if [ "$v" -lt 0 ]; then echo 0; return; fi
  if [ "$v" -gt 100 ]; then echo 100; return; fi
  echo "$v"
}

MOVIES_MIGRATION_PERCENT="$(clamp "$MOVIES_MIGRATION_PERCENT")"
EVENTS_MIGRATION_PERCENT="$(clamp "$EVENTS_MIGRATION_PERCENT")"

if [ "$GRADUAL_MIGRATION" = "true" ]; then
  MOVIES_NEW_WEIGHT="$MOVIES_MIGRATION_PERCENT"
  MOVIES_OLD_WEIGHT="$((100 - MOVIES_MIGRATION_PERCENT))"

  EVENTS_NEW_WEIGHT="$EVENTS_MIGRATION_PERCENT"
  EVENTS_OLD_WEIGHT="$((100 - EVENTS_MIGRATION_PERCENT))"
else
  # миграция выключена => всё в монолит
  MOVIES_NEW_WEIGHT="0"
  MOVIES_OLD_WEIGHT="100"
  EVENTS_NEW_WEIGHT="0"
  EVENTS_OLD_WEIGHT="100"
fi

export MOVIES_NEW_WEIGHT MOVIES_OLD_WEIGHT EVENTS_NEW_WEIGHT EVENTS_OLD_WEIGHT

# генерим итоговый declarative config
envsubst < /etc/kong/kong.yml.template > /etc/kong/kong.yml

exec "$@"
