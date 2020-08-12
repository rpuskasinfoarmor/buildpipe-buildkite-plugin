#!/usr/bin/env bats

load "$BATS_PATH/load.bash"

setup() {
  _GET_CHANGED_FILE='log --name-only --no-merges --pretty=format: origin..HEAD'
  stub git "${_GET_CHANGED_FILE} : echo 'deploy/thing.txt'"
  stub buildkite-agent pipeline upload
}

teardown() {
  unstub git
  # TODO: fix not being able to unstub
  # unstub buildkite-agent
}


@test "Checks projects affected" {
  export BUILDKITE_PLUGIN_BUILDPIPE_DYNAMIC_PIPELINE="tests/distinct_pipeline.yml"
  export BUILDKITE_PLUGIN_BUILDPIPE_LOG_LEVEL="DEBUG"
  export BUILDKITE_PLUGIN_BUILDPIPE_TEST_MODE="true"
  export BUILDKITE_BRANCH="not_master"

  run hooks/command

  assert_success
  IFS=''
  while read line
  do
    assert_line "$line"
  done << EOM
steps:
  - label: package
    env:
      BUILDPIPE_SCOPE: distinct
    command:
      - make package
    agents:
      - queue=build
  - wait
  - label: deploy
    command:
      - make deploy
    env:
      BUILDPIPE_SCOPE: distinct
  - wait
  - label: test
    env:
      BUILDPIPE_SCOPE: distinct
    command:
      - make test
EOM
}
