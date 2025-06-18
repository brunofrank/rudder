![Rudder](https://user-images.githubusercontent.com/40139/110044103-5220a600-7d27-11eb-84b7-0d462159a4f9.png)

# Rudder

rudder is a set of helpers to easily opperate docker-compose commands

# Installation

```sh
sh -c "$(curl -fsSL https://raw.githubusercontent.com/brunofrank/rudder/refs/heads/main/install.sh)"
```

In the project root you want to use rudder run:

```sh
rudder --init
```

# Project configuration

Given that every project has diferent docker-compose you can teach rudder your comands using
the `.rudder.yml` file in the root of you project.

Following a example of a Rails application:

```yaml
rudder:
  default_service: web
  commands:
    ssh: bash -l
    bundle: bundle $ARGS
    rails: bundle exec rails $ARGS
    db:migrate: rails db:migrate
    db:rollback: rails db:migrate
    rake: bundle exec rake $ARGS
    gem: bundle exec gem $ARGS
    console: bundle exec rails c
    logs: logs -f @ host # You use @ host do run the command in the host machine
    yarn: yarn $ARGS @ frontend # You use @ to define in what docker compose service it should run
    restart: restart @ host
    restart:web: restart web @ host
    pristine:
      - echo "This will destroy your containers and replace them with new ones." @ host
      - docker compose down -v @ host
      - docker compose up --build --force-recreate --no-start @ host
      - yarn install
      - bundle
      - docker compose restart @ host
      - echo "Creating data..." @ host
      - rake db:create
      - rake db:extensions
      - rake db:schema:load
      - rake db:migrate
      - rake db:seed
      - echo "Creating data... Done! ;)" @ host
      - restart @ host
      - echo "It may take few minutes to launch all containers." @ host
      - echo "You can access your environment at https://demo.lvh.me:3000" @ host
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

# Updating

To update rudder you just need to run:

```sh
rudder --update
```
