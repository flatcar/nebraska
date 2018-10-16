# Deploying rollerd for testing on local computer

- go to https://smee.io/ and press the `Start a new channel` button,
  you'll get redirected to something like
  `https://smee.io/asfdFA7B87SD98F7`

- on the computer create some directory for nodejs stuff (say
  `~/nodejsstuff`), then in that directory do `npm install
  smee-client`, after that run `~/nodejsstuff/node_modules/.bin/smee
  -u https://smee.io/asfdFA7B87SD98F7 -p 8000 -P /login/webhook`

- now you need a github app, go to
  `https://github.com/settings/apps/new` (or, if doing it in some
  organization, to
  `https://github.com/organizations/<ORG>/settings/apps/new`) and fill
  the following fields:

  - `GitHub App name` - just put some fancy name, must be unique I
    think

  - `Homepage URL` - `https://smee.io/asfdFA7B87SD98F7` (or whatever
    you've got, not really important)

  - `User authorization callback URL` - `http://localhost:8000/login/cb`

  - `Webhook URL` - `https://smee.io/asfdFA7B87SD98F7`

  - `Webhook secret` - another secret stuff, this time for webhook
    message validation

  - Enable SSL verification

  - `Permissions` - `Access: Read-only` to `Organization members`

  - `User permissions` - none needed

  - `Subscribe to events` - tick `Membership`, `Organization` and `Team`

  - `Where can this GitHub App be installed?` - `Only on this account`

- Press `Create GitHub App` button

- Next thing you'll get is `OAuth credentials` at the bottom of the
  page of the app you just created, we will need both `Client ID` and
  `Client secret`

- You also need to install the app you just created

  - go to `https://github.com/settings/apps` (or, in case of
    organization,
    `https://github.com/organizations/<ORG>/settings/apps`)

  - Click `Edit` button for your new app

  - Choose `Install App` on the left of the page and perform the
    installation

- now go to the coreroller project directory and run `make all`

- start the database: `docker run --privileged -d -p
  127.0.0.1:5432:5432 coreroller/postgres`

- make sure that there is a coreroller team in the coreroller database
  that is named after a github organization or a github team members
  of which should be allowed to manage the application

  - naming of the coreroller teams is `<ORG>` for coreroller teams
    based on github organizations or `<ORG>/<TEAM>` for coreroller
    teams based on github teams

  - you can use `userctl` tool to create/rename coreroller teams; the
    tool is built as a part of `make tools`

  - use `userctl -list-teams` to see teams

  - use `userctl -add-team <TEAM_NAME>` to create a new team

  - use `userctl -change-team-name-to <TEAM_NAME> -add-team
    <OLD_TEAM_NAME>` to rename `<OLD_TEAM_NAME>` to `<TEAM_NAME>`

    - yeah, I know, the `-add-team` part here is confusing

- start the coreroller backend:

  - `LOGXI=* rollerd -client-id <CLIENT_ID> -client-secret
    <CLIENT_SECRET> -http-static-dir $PWD/frontend/built -http-log
    -webhook-secret <WEBHOOK_SECRET>`

    - run `rollerd -h` to learn about env vars you can use instead of
      flags

- in the browser, access `http://localhost:8000`

# Deploying on some server

Let's assume that our rollerd instance is reachable through the
`coreroller.example.com` address. In this case we can just drop this
`smee` and `localhost:8000` nonsense. So the following fields in the
app configuration can have different values:

  - `Homepage URL` - `https://coreroller.example.com` (or whatever
    you've got, not really important)

  - `User authorization callback URL` - `https://coreroller.example.com/login/cb`

  - `Webhook URL` - `https://coreroller.example.com/login/webhook`

Rest of the steps is the same.

# Personal access tokens

How `rollerd` instance does authentication depends on existence of the
`Authorization` header in the first request. Basically if the header
does not exist then `rollerd` will go with the "Login through GitHub"
route, which means redirecting to github, logging in there if you
weren't logged in before, authorizing the app if it wasn't authorized
before. This is not exacltly friendly for some services (for example,
prometheus won't work with that).

If the `Authorization` header exists, it must be like `Authorization:
bearer <TOKEN>`. The `<TOKEN>` part should be personal access token
generated on github (`https://github.com/settings/tokens`).

Personal access token requires just one scope: `read:org`.

# Troubleshooting:

- I'm getting a blank page!

  - You likely visited coreroller frontend website before, so browser
    likely has cached the `index.html` page, so it won't get it from
    `rollerd`, but instead start asking for some CSS and javascript
    stuff outright, which it won't get. That results in a blank
    page. Force the browser to get `index.html` from `rollerd` by
    cleaning the cache for localhost (or the server where the
    `rollerd` instance is deployed). I suppose we should fix `rollerd`
    to serve `index.html` with some `Cache-Control: no-store` header
    in response.
