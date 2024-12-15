# Description:
# This script does the following:
# - Make sure the working directory is clean.
# - Pushes a new version of the app (if it does not already exist; uses git sha as version).
# - Updates the candidate instance to use the new version.
# - Runs an acceptance test between the candidate and baseline instances.
# - Waits for the test to complete.
# - Posts the result to Slack (if requested).

import os
import subprocess
from datetime import datetime, timezone

from nextmv import cloud

APP_ID = "nextroute-bench"
API_KEY = os.environ["NEXTMV_API_KEY"]


METRICS = [
    cloud.Metric(
        field="result.value",
        metric_type=cloud.MetricType.direct_comparison,
        params=cloud.MetricParams(
            tolerance=cloud.MetricTolerance(
                value=0.05,
                type=cloud.ToleranceType.relative,
            ),
            operator=cloud.Comparison.less_than_or_equal_to,
        ),
        statistic=cloud.StatisticType.mean,
    )
]


def ensure_clean_working_directory():
    """
    Ensure the working directory is clean by throwing an exception if it is not.
    """
    if os.system("git diff --quiet") != 0 or os.system("git diff --cached --quiet") != 0:
        raise Exception("Working directory is not clean")


def get_tag(app: cloud.Application) -> str:
    """
    Get the tag for the new version.
    If the version already exists, we append a timestamp to the tag.
    """
    # If we don't have a tag, we generate one based on the git sha and a timestamp
    git_sha = subprocess.check_output(["git", "rev-parse", "HEAD"]).decode().strip()[0:8]
    # If the version already exists, we append a timestamp to the tag
    exists = False
    try:
        app.version(git_sha)
        exists = True
    except Exception:
        pass
    if exists:
        ts = (
            datetime.now(timezone.utc)
            .replace(microsecond=0)
            .isoformat()
            .replace("+00:00", "Z")
            .replace(":", "")
            .replace("-", "")
        )
        return f"{git_sha}-{ts}"
    # Otherwise, we just use the git sha
    return git_sha


def push_new_version(app: cloud.Application, tag: str) -> None:
    """
    Push a new version of the app and update the candidate instance to use it.
    """
    app.push(app_dir=".")
    version_id = f"auto-{tag}"
    app.new_version(
        id=version_id,
        name=f"Auto version {tag}",
        description=f"Automatically generated version {tag}",
    )
    app.update_instance(
        id="candidate",
        version_id=version_id,
    )


def run_acceptance_test(
    app: cloud.Application,
    tag: str,
) -> cloud.AcceptanceTest:
    """
    Run an acceptance test between the candidate and baseline instances.
    """
    id = f"auto-{tag}"
    print(f"Running acceptance test with ID: {id}")
    print("Waiting for the test to complete...")
    result = app.new_acceptance_test_with_result(
        candidate_instance_id="candidate",
        baseline_instance_id="baseline",
        id=id,
        metrics=METRICS,
        name=f"Auto-test {tag}",
        description=f"Automated test for {tag}",
        input_set_id="nextroute-bench-v20",
        polling_options=cloud.PollingOptions(
            max_duration=600,  # 10 minutes
            max_tries=1000,  # basically forever - we'll stop by duration
        ),
    )
    passed = "unknown"
    if result and result.results:
        passed = "passed" if result.results.passed else "failed"
    print(f"Acceptance test completed with status: {passed}")
    return result


def main():
    """
    Main function that runs the benchmark.
    """
    # Change to the directory of the app (sibling directory of this script)
    os.chdir(os.path.join(os.path.dirname(__file__), "..", "cmd"))

    ensure_clean_working_directory()

    client = cloud.Client(api_key=os.getenv("NEXTMV_API_KEY"))
    app = cloud.Application(client=client, id=APP_ID)

    tag = get_tag(app)

    push_new_version(app, tag)

    run_acceptance_test(app, tag)


if __name__ == "__main__":
    main()
