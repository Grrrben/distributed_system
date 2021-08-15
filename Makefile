build_log_service:
	go build ./cmd/logservice
build_registry_service:
	go build ./cmd/registryservice
build_grading_service:
	go build ./cmd/gradingservice
build_teacherportal:
	go build -o portal ./cmd/teacherportal

