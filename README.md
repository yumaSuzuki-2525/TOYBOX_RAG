
### Docker
```docker
docker-compose up -d

# 完全削除
# docker-compose down -v
```

### AIモデルのダウンロード
```cmd
# 1. 回答生成用の Llama 3 をダウンロード（数GB:数分かかります）
docker exec -it ollama_server ollama pull llama3

# 2. ベクトル化用のモデルをダウンロード
docker exec -it ollama_server ollama pull nomic-embed-text
```

### DB用意
```postgres
CREATE EXTENSION vector;
-- 1. ベクトル拡張機能を有効化
CREATE EXTENSION IF NOT EXISTS vector;

-- 2. ドキュメント保存用テーブルの作成
CREATE TABLE IF NOT EXISTS documents (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,         -- ドキュメントの文章（チャンク）
    -- nomic-embed-text は768次元のベクトルを出力するため 768 を指定
    embedding vector(768) NOT NULL 
);

-- 3. 検索を高速化するためのインデックス
CREATE INDEX ON documents USING hnsw (embedding vector_cosine_ops);
```


### ローカルで動かす際の.envファイル
```.env
# DB設定
DB_HOST=localhost
DB_PORT=5434
DB_USER=postgres
DB_PASS=postgres
DB_NAME=postgres
# OllamaURL
OLLAMA_URL=http://localhost:11434/api
# ファイル読み込み/出力
SOURCE_DIR=./public/docs
PROCESSED_DIR=./public/processed
```




### TODO

> * 0001: 現在検索でSELECT content FROM documents ORDER BY embedding <-> $1 LIMIT 1;のlimit 1だが複数取得しないと情報が抜け落ちているような気がする。複数行取得するかして情報を多くAIに渡せれるようにする。

> * 0002: ベクトル検索は曖昧検索で完全一致検索が苦手とのこと。質問内容から形態素解析で名詞取得するか、入力時""で囲うかしてキーワード検索ができるようにする。ハイブリッド検索。

> * 0003: Ollama, Llamaを使っているが、アップデートができないと今回限りになるので、他のAIも使えるように調整。その際、Postgresqlのデータを使っているが削除するか、パーティショニングするか、DBを切り替えるかなどの方法も検討する。

> * 0004: 現状Docker起動のCLIだが、ローカルLLMを使っているのであればC#の.Netで呼び出せる用にする？その場合サービス起動しておく必要があるか要検討。現状Docker起動するとcpuを結構使っている。

> * 0005: ローカルLLMなのでセキュリティは担保されている。DB登録時ジャンル分けできればDBにただ流し込んでいる時よりもクエリのパフォーマンスは上がると思う。ジャンルテーブル作って管理？

> * 0006: 100ページあるPDFなどを探す。insertがN+1なのでCOPYのバルクインサート検討。

> * 0007: 不要な情報も必ずあるがどう判断する？ただ流し込むと容量が圧迫される。