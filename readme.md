# watchbot

grocery shops list parsed:
- arbuz.kz (web+jwt+api)
- a-store (instashop)
- galmart (instashop)
- Metro Cash & Carry (instashop)
- Interfood (instashop)
- Зеленый базар (instashop)
- Airba fresh (technodom)

todo:
- magnum (app)
- small (no app, no website)
- use proxy list 
- proxy rotation
- gui

create migration:
migrate create -ext sql -dir ./migrations/ -seq migration_name

run migration
docker-compose up watchbot_migrate