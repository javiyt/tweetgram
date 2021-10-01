# Tweetgram

[![codecov](https://codecov.io/gh/javiyt/tweetgram/branch/main/graph/badge.svg?token=Q15YVM2SMC)](https://codecov.io/gh/javiyt/tweetgram)
[![Test](https://github.com/javiyt/tweetgram/actions/workflows/ci.yml/badge.svg)](https://github.com/javiyt/tweetgram/actions/workflows/ci.yml)

Telegram bot to publish post to twitter

## How to set up the environment?

First you need to install, if not installed on your machine, [Task](https://taskfile.dev/#/). Once installed you only
need to run in your command line:

```
task setup
```

And you are done, it will generate the env files needed to run the bot and will download all dependencies, also run all
code generation tools.

###  Configuring the bot

You can find a file inside cmd folder called env where you can adjust the variables, take into account the generated one
are random values. The content of the env file will look like following:

```
BOT_TOKEN=1234567890:G8o4ATpRsfUtl0p7N1HW9S2IdIcxSRoSY67
ADMINS=123456789
BROADCAST_CHANNEL=-1234567890123
TWITTER_API_KEY=c7FU8EvL9smKN2k2IN0yur67k
TWITTER_API_SECRET=LYzF53kJoVK46rp859rQCw6Dqw6TpQV668aemPb2KI9GUxTTU0
TWITTER_BEARER_TOKEN=hIlQI351HEPT6xbA4xHnYRfgOsF8jqcPT5m6Ec0VeCXtUyOY9Mzy6uFYevH%4ys86GL3KfO1ZRBwichZOlGDYyZ52Ht2BXh2WgUFvywJKbRq9lMH
TWITTER_ACCESS_TOKEN=123456789-0xW7uNw2mEykTVTTKHS32y3oIXHab5hSh7POa0Wf
TWITTER_ACCESS_SECRET=BWa9T8hkEEj5yCutPwJTs7Vk4f1wfj690Dq3UGCyf9YQB
ENVIRONMENT=TEST
LOG_FILE=/var/log/tweetgram.log
```
Env file variables are self-explanatory

Check env.test file, you only need there all the variables that should be overridden in order to run a test instance of
the bot. Take into account env.test file is not needed to run the test case they set up the appropriate variables to run
them. Remove all not needed variables from env.test file

### Running a local instance of the bot

You just need to run the following command:

```
task run-test
```

Remember all env variables will be overridden by the ones defined in env.test

## Deploying the bot
There's an action called deploy that you can trigger to deploy the bot. Some [secrets should be added](https://docs.github.com/en/actions/security-guides/encrypted-secrets) to your account before running the deployment script. The variables that should be added are:
| Variable      | Description                                  |
|---------------|----------------------------------------------|
| ENV_FILE      | cmd/env file content                         |
| ENV_TEST_FILE | cmd/env.test file content                    |
| HOST          | Host of the server to deploy the bot         |
| USERNAME      | User to connect to the server                |
| SSHKEY        | SSH key of the user to login onto the server |
| PORT          | Port where the SSH server is running         |
| PASSPHRASE    | Passphrase to decrypt the SSH key            |
| FOLDER        | Folder to deploy the bot binary              |
| BINARY_NAME   | Name for the generated binary                |