Create git
0. pwd -> must be in root directory
1. git init
2. git add .
3. git commit -m "Initial commit"

Then go into github  and create new repository 
and 
…or push an existing repository from the command line
1. git remote add origin git@github.com:bojkrstic/testkrle.git
2. git branch -M main
3. git push -u origin main

## Local development

Before pokretanja servera postavi konekcioni string u promenljivu okruženja, na primer:

Promenljivu možeš setovati direktno u terminalu:

```bash
export DB_DSN="root:root@tcp(127.0.0.1:3308)/bulk_gate?parseTime=true"
```

ili napravi `.env` fajl u root direktorijumu sa sadržajem:

```
DB_DSN=root:root@tcp(127.0.0.1:3308)/bulk_gate?parseTime=true
```

Server će pri startu automatski pokupiti vrednost iz `.env` ako promenljiva okruženja ne postoji.

Zatim pokreni aplikaciju:

```bash
go run ./cmd/web
```
