test:
	go install github.com/zoncoen/scenarigo/cmd/scenarigo@v0.17.1
	scenarigo run -c ./scenarigo/scenarigo.yaml
	jq '.files[0].scenarios[0].steps | map({response: .logs.info[1]})' scenarigo/output.json > result.json
# P=.env.PROJECT_ID
gen-meta:
	bq show --format=prettyjson ${P}:metadata.users > info.json
	cat ./info.json | jq '{tableReference, description, schema}' > schema.txt
	rm -rf info.json