# kfc-reactor
redisに登録したキーワードを含む文字列がSlackに投稿されたら `:kfc:` リアクションを付けるアプリ

# 要求事項
Slackに `kfc` という名前の絵文字が登録されていること

# 動かしかた
1. Slack Appを登録
    1. 任意の名前のアプリを作成
    2. Slash Commands, Event Subscriptions, Bots, PermissionsをOnにする
    3. 権限を設定する
        * Bot Token Scopes: app_mentions:read, channels:history, channels:join, channels:read, chat:write, commands, emoji:read, reactions:write
        * User Token Scopes: channels:history, channels:read, emoji:read, reactions:write, search:read
    4. Install Apps

2. Herokuにデプロイする
    * [![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy)
    * Heroku Redisアドオンを追加する

3. Herokuの環境変数をセットしておく
    * `SLACK_BOT_TOKEN` : 登録したSlackアプリのBot User OAuth Access Token
    * `SLACK_USER_TOKEN` : 登録したSlackアプリのOAuth Access Token
    * `KFC_REACTOR_SIGNING_SECRET` : 登録したSlackアプリのSigning Secret
    * `PORT` : listenするポート番号
    * `KFC_REACTOR_HOME_CHANNEL_ID` : 覚えたキーワードを通知するためのチャンネルのID
    * `REDIS_URL` : RedisのURL

4. Slack AppのEvent Subscriptions > Request URLにHerokuで公開したURLを入力し、Verifiedのチェックが付くことを確認する

5. Slack AppのSlash Commandsの設定でRequest URLに「Herokuで公開したURL/command」を入力して保存する

6. redisにキーワードを登録する
    * `heroku redis:cli [redis instance] --confirm kfc-reactor`
    * `set keyword:[keyword] null`
