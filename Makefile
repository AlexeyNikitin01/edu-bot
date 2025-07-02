pg:
	docker run --name edu -e POSTGRES_PASSWORD=pass -e POSTGRES_USER=postgres -e POSTGRES_DB=edu -p 7878:5432 -d postgres

up:
	pgmigrate -c config.yml -o up

down:
	pgmigrate -c config.yml -o down

reset:
	pgmigrate -c config.yml -o reset

entity:
	sqlboiler psql -c etc/config.yml -p edu -o internal/repo/edu --add-soft-deletes

ssh:
	ssh -i home/.ssh/id_ed25519 -l aleksey 51.250.98.75
