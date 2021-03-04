# rudder

rudder is a set of helpers to easily opperate docker-compose commands

# Project configuration

Given that every project has diferent docker-compose you can teach rudder your comands using
the `.rudder` file in the root of you project.

Following a example of a Rails application:

```yaml
rudder:
  default_service: web
  commands:
    ssh: bash -l
    bundle: bundle $@
    rails: bundle exec rails $@
    db:migrate: rails db:migrate
    db:rollback: rails db:migrate
    rake: bundle exec rake $@
    gem: bundle exec gem $@
    console: bundle exec rails c
    logs: host:docker-compose logs -f
    yarn: yarn $@
    restart: host:docker-compose restart
    restart:web: host:docker-compose restart web
    pristine:
      - host:echo "This will destroy your containers and replace them with new ones."
      - host:docker-compose down -v
      - host:docker-compose up --build --force-recreate --no-start
      - yarn install
      - bundle
      - host:docker-compose restart
      - host:echo "Creating data..."
      - rake db:create
      - rake db:extensions
      - rake db:schema:load
      - rake db:migrate
      - rake db:seed
      - host:echo "Creating data... Done! ;)"
      - host:docker-compose restart
      - host:echo "It may take few minutes to launch all containers."
      - host:echo "You can access your environment at https://demo.lvh.me:3000"
    setup:
      - yarn install
      - bundle
      - rake db:create
      - rake db:extensions
      - rake db:schema:load
      - rake db:migrate
      - rake db:seed
    guard: bundle exec guard
```
