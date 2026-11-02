.PHONY: check check-backend check-frontend check-migrations

# Same entrypoint used by Claude hooks and CI.
check:
	scripts/check.sh all

check-backend:
	scripts/check.sh backend

check-frontend:
	scripts/check.sh frontend

check-migrations:
	scripts/check.sh migrations
