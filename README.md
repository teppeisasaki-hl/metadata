## メタデータをプロンプトに含めることでレビューの精度がどう変わるのかを検証するリポジトリです。

* claude 3.5 sonnet を使用して下記の複雑なテーブルを作成しました

```
CREATE TABLE `metadata.users` (
  id INT64 NOT NULL OPTIONS (description='ユーザーの識別子'),
  name STRING OPTIONS (description='ユーザー名'),
  birth_date DATE OPTIONS (description='ユーザーの生年月日'),
  contact_info STRUCT<
    primary_email STRING OPTIONS (description='プライマリーメールアドレス'), 
    secondary_email STRING OPTIONS (description='セカンダリーメールアドレス'), 
    phone STRING OPTIONS (description='電話番号を含む')
  > OPTIONS (description='ユーザーの連絡先情報'),
  account_details STRUCT<
    creation_timestamp TIMESTAMP OPTIONS (description='作成日時'),
    last_activity TIMESTAMP OPTIONS (description='最終アクティビティ'),
    is_premium BOOL OPTIONS (description='プレミアムステータス'),
    subscription_type STRING OPTIONS (description='サブスクリプションタイプを含む')
  > OPTIONS (description='アカウントの詳細情報。'),
  metrics STRUCT<
    engagement_score FLOAT64 OPTIONS (description='エンゲージメントスコア'),
    lifetime_value NUMERIC OPTIONS (description='ライフタイムバリュー'),
    average_session_duration INT64 OPTIONS (description='平均セッション時間を含む')
  > OPTIONS (description='ユーザーに関する各種指標。'),
  preferences ARRAY<STRUCT<
    category STRING OPTIONS (description='カテゴリ'),
    setting STRING OPTIONS (description='設定項目'),
    value STRING OPTIONS (description='値')
  >> OPTIONS (description='ユーザーの設定。'),
  activity_log ARRAY<STRUCT<
    action STRING OPTIONS (description='アクション'),
    timestamp TIMESTAMP OPTIONS (description='タイムスタンプ'),
    details STRING OPTIONS (description='詳細情報')
  >> OPTIONS (description='ユーザーの活動ログ。'),
  geolocation GEOGRAPHY OPTIONS (description='ユーザーの地理的位置情報'),
)
OPTIONS(
  description="このテーブルは複雑な構造を持つ類似カラムのユーザー情報を格納します。各カラムには詳細な説明が付いています。"
);
```

## 使用方法
* 起動
    * `docker compose up`
* テスト
    * `make test`

## シナリオファイル
* metadata.yaml の使用方法
    * with
        * none: 何もなし
        * metadata: prompt に作成したテーブルのメタ情報が追加されます (カラム名, 型, 説明)
    * sql
        * sql を書いてください
    * object
        * sql がこのパラメーターに沿ったものになっているかを判定します