# kfc-reactor
Slackに「昼」「lunch」などを含む文字列が投稿されたら `:kfc:` リアクションを付けるアプリ

# 要求事項
Slackに `kfc` という名前の絵文字が登録されていること

# 動かしかた
1. Slack Appを登録
    1. 任意の名前のアプリを作成
    2. Event Subscriptionsの設定でEnable EventsをOnにする
    3. Subscribe to events on behalf of usersでmessage.channelsを選択
    4. OAuth & Permissions内のUser Token Scopesでchannels:history, emoji:read, reactions:writeを選択
    5. Install App

2. 環境変数をセットしておく
    * `SLACK_USER_TOKEN` : 登録したSlackアプリのOAuth Access Token
    * `KFC_REACTOR_SIGNING_SECRET` : 登録したSlackアプリのSigning Secret
    * `PORT` : listenするポート番号

3. ngrokで外部公開する、herokuにデプロイするなどして実行する
    * 公開したURLをSlackアプリ設定画面のEvent Subscriptions内、Request URL欄に入力し、Verifiedのチェックが付くことを確認する
    * Save Changes

4. 以上。認証したユーザが参加しているチャンネルで「昼」「lunch」等を含む文字列が投稿されると、認証したユーザとしてその投稿に `:kfc:` リアクションを付ける。
