run:
	go run cmd/film-sync/main.go

seed:
	go run scripts/seed/seed.go

# Resend Discord success DM for a saved scan (dev server must be running).
# Usage: make notify EMAIL=19f7736cfa68437f
#    or: make notify SCAN=<mongodb_object_id>
notify:
	@if [ -n "$(EMAIL)" ]; then \
		curl -fsS "http://localhost:$${PORT:-3001}/notify?email_id=$(EMAIL)"; \
	elif [ -n "$(SCAN)" ]; then \
		curl -fsS "http://localhost:$${PORT:-3001}/notify?scan_id=$(SCAN)"; \
	else \
		echo "Usage: make notify EMAIL=<gmail_message_id>   OR   make notify SCAN=<scan_object_id>"; \
		exit 1; \
	fi
