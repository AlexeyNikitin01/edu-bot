pg:
	docker run --name edu -e POSTGRES_PASSWORD=pass -e POSTGRES_USER=postgres -e POSTGRES_DB=edu -p 7878:5432 -d postgres
