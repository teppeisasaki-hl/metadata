package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/vertexai/genai"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

var (
	logger, _             = zap.NewProduction()
	promptWithoutMetadata = `あなたは SQL のレビュワーです。

# 指示:
- 入力 SQL が SQL の目的に合っているか判断しなさい。
- その後目的と違う場合や他に適したカラムがある場合指摘してください。

# SQL の目的
- %s

# 入力 SQL : """
%s
"""

# 出力形式:
- Markdown
`
	promptWithMetadata = `あなたは SQL のレビュワーです。

# 指示:
- 入力 SQL が SQL の目的に合っているか判断しなさい。
- その後テーブルメタデータを参考にして、目的と違う場合や他に適したカラムがある場合指摘してください。

# SQL の目的
- %s

# 入力 SQL : """
%s
"""

# テーブルメタデータ:
%s

# 出力形式:
- Markdown
`
	promptWithSchemaFile = `あなたは SQL のレビュワーです。

# 指示:
- 入力 SQL が SQL の目的に合っているか判断しなさい。
- その後テーブルメタデータの概要と実データを参考にして、目的と違う場合や他に適したカラムがある場合指摘してください。

# SQL の目的
- %s

# 入力 SQL : """
%s
"""

# テーブルメタデータ概要 :
- tableRefrence の中にはデータセットとプロジェクトとテーブルの ID が含まれます。
- description にはテーブルの説明が記述されています。
- shcema の中には fields が存在し、その中にはカラムの情報が記述されています。
- カラムの情報には name: 名前, type: データ型, description: カラムの説明 などが含まれネストされたフィールドも存在します。

# 実際のテーブルメタデータ:
%s


# 出力形式:
- Markdown
`
)

type MetadataResponse struct {
	Prompt string
	Output string
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", handler)
	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), r)
}

func handler(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		obj  = r.URL.Query().Get("object")
		sql  = r.URL.Query().Get("sql")
		with = r.URL.Query().Get("with")
	)

	client, err := genai.NewClient(ctx, os.Getenv("PROJECT_ID"), "asia-northeast1", option.WithCredentialsFile("./sa.json"))
	if err != nil {
		errorRes(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	prompt := ""
	if with == "metadata" {
		ap, err := genMetadata(ctx)
		if err != nil {
			errorRes(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		prompt = fmt.Sprintf(promptWithMetadata, obj, sql, ap)
	} else if with == "schema_file" {
		c, err := os.ReadFile("schema.txt")
		if err != nil {
			errorRes(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		prompt = fmt.Sprintf(promptWithSchemaFile, obj, sql, string(c))
	} else {
		prompt = fmt.Sprintf(promptWithoutMetadata, obj, sql)
	}

	genModel := client.GenerativeModel("gemini-1.0-pro")
	genModel.Temperature = genai.Ptr[float32](0)
	resp, err := genModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		errorRes(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	var result string
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				textPart, ok := part.(genai.Text)
				if ok {
					result += string(textPart)
				}
			}
		}
	}

	success(w, r, &struct {
		Prompt string
		Output string
	}{prompt, result})
}

func genMetadata(ctx context.Context) (string, error) {
	tmp := ""
	c, err := bigquery.NewClient(ctx, os.Getenv("PROJECT_ID"), option.WithCredentialsFile("./sa.json"))
	if err != nil {
		return "", err
	}
	defer c.Close()

	md, err := c.Dataset(os.Getenv("DATASET")).Table(os.Getenv("TABLE")).Metadata(ctx)
	if err != nil {
		return "", err
	}

	for i, s := range md.Schema {
		if i == 0 {
			tmp += fmt.Sprintf("\n\n# テーブルメタデータ \n テーブル名: %s \n", os.Getenv("TABLE"))
		}
		tmp += fmt.Sprintf("%s. %s ( %s ): %s \n", strconv.Itoa(i+1), s.Name, s.Type, s.Description)
		for a, cs := range s.Schema {
			tmp += fmt.Sprintf("  %s. %s ( %s ): %s \n", strconv.Itoa(a+1), cs.Name, cs.Type, cs.Description)
		}
	}

	return tmp, nil
}

func success(w http.ResponseWriter, r *http.Request, d interface{}) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, d)
}

func errorRes(w http.ResponseWriter, r *http.Request, status int, d string) {
	render.Status(r, status)
	render.JSON(w, r, map[string]string{"message": d})
}
