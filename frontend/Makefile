.PHONY: dev install

install:
	npm install

dev: install
	npm run dev -- --port 5173

wait:
	@echo "Waiting for http://localhost:5173 to be ready..."
	@until curl -sf http://localhost:5173 > /dev/null; do sleep 1; done
	@echo "Frontend is up at http://localhost:5173"
