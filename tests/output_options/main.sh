go run . \
    -solve.duration 10s \
    -format.disable.progression \
    -solve.parallelruns 1 \
    -solve.iterations 50 \
    -solve.rundeterministically \
    -solve.startsolutions 1 \
    -runner.input.path ../golden/testdata/template_input.json 2>/dev/null | jq ".options" # Silence bunny output, since we're only interested in the options
