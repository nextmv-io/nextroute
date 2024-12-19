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

import requests
from nextmv import cloud

APP_ID = "nextroute-bench"
API_KEY = os.environ["BENCHMARK_API_KEY_PROD"]
SLACK_WEBHOOK = os.getenv("SLACK_URL_DEV_SCIENCE", None)
ACCOUNT_ID = os.getenv("BENCHMARK_ACCOUNT_ID", None)
BRANCH_NAME = os.getenv("BRANCH_NAME", None)


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
        statistic=cloud.StatisticType.shifted_geometric_mean,
    )
]


def ensure_clean_working_directory():
    """
    Ensure the working directory is clean by throwing an exception if it is not.
    """
    if os.system("git diff --quiet") != 0 or os.system("git diff --cached --quiet") != 0:
        raise Exception("Working directory is not clean")


def get_id(app: cloud.Application) -> tuple[str, str]:
    """
    Get the ID for the new version (and just the tag).
    If the version already exists, we append a timestamp to the ID.
    """
    # Create ID based on git sha.
    tag = subprocess.check_output(["git", "rev-parse", "HEAD"]).decode().strip()[0:8]
    version_id = f"auto-{tag}"
    # If the version already exists, we append a timestamp to the ID.
    exists = False
    try:
        app.version(version_id)
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
        version_id = f"{version_id}-{ts}"
        tag = f"{tag}-{ts}"
    # Otherwise, we just use the git sha.
    return version_id, tag


def push_new_version(app: cloud.Application, tag: str) -> None:
    """
    Push a new version of the app and update the candidate instance to use it.
    """
    app.push(app_dir=".")
    app.new_version(
        id=tag,
        name=f"Auto version {tag}",
        description=f"Automatically generated version {tag}",
    )
    instance = app.instance("candidate")
    app.update_instance(
        id="candidate",
        version_id=tag,
        name=instance.name,  # Name is required, but we don't want to change it
    )


def upgrade_baseline(app: cloud.Application, version_id: str) -> None:
    """
    Upgrade the baseline instance to use the new version.
    """
    instance = app.instance("baseline")
    app.update_instance(
        id="baseline",
        version_id=version_id,
        name=instance.name,  # Name is required, but we don't want to change it
    )


def run_acceptance_test(
    app: cloud.Application,
    id: str,
    tag: str,
) -> cloud.AcceptanceTest:
    """
    Run an acceptance test between the candidate and baseline instances.
    """
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
    return result


def create_test_url(result_id: str) -> str:
    """
    Create a URL to the acceptance test result.
    """
    if ACCOUNT_ID:
        return f"https://cloud.nextmv.io/acc/{ACCOUNT_ID}/app/nextroute-bench/experiment/acceptance/{result_id}"
    return "unavailable"


def write_to_summary(content):
    """Appends content to the GitHub Actions step summary (if available)."""
    summary_file = os.getenv("GITHUB_STEP_SUMMARY")
    if not summary_file:
        return

    # Write content to the summary file
    with open(summary_file, "a") as f:
        f.write(content + "\n")


def main():
    """
    Main function that runs the benchmark.
    """
    # Change to the directory of the app (sibling directory of this script)
    os.chdir(os.path.join(os.path.dirname(__file__), "..", "cmd"))

    print("Making sure the working directory is clean...")
    ensure_clean_working_directory()

    client = cloud.Client(api_key=API_KEY)
    app = cloud.Application(client=client, id=APP_ID)

    id, tag = get_id(app)  # id is used as version and acceptance test ID

    print(f"Pushing new version with ID: {id}")
    push_new_version(app, id)

    write_to_summary("# Acceptance Test Report")
    write_to_summary("")
    write_to_summary(f"ID: {id}")
    url = create_test_url(id)
    write_to_summary(f"Link: [link]({url})")
    print(f"::notice::Acceptance test URL: {url}", flush=True)

    print(f"Running acceptance test with ID: {id}")
    print("Waiting for it to complete...")
    result = run_acceptance_test(app, id, tag)
    passed = "unknown"
    if result and result.results:
        passed = "passed" if result.results.passed else "failed"
    print(f"Acceptance test completed with status: {passed}")

    if SLACK_WEBHOOK and BRANCH_NAME == "develop":
        print("Posting to Slack...")
        response = requests.post(
            SLACK_WEBHOOK,
            json={
                "text": f"nextroute acceptance test {result and result.id} completed with status: {passed}"
                + f" (<{create_test_url(result and result.id)}|View results>)",
            },
        )

        if response.status_code != 200:
            print(f"Failed to send notification to Slack: {response.text}")
        else:
            print("Notification sent to Slack")

    write_to_summary("")
    write_to_summary(f"Result: {passed}")
    if result and result.results:
        if result.results.error:
            write_to_summary(f"Error: {result.results.error}")
        else:
            write_to_summary("Metrics:")
            write_to_summary("")
            for metric in result.results.metric_results:
                write_to_summary(f"- {metric.metric.field}: {metric.passed}")

    if BRANCH_NAME == "develop":
        print("Upgrading baseline instance to use the new version...")
        upgrade_baseline(app, id)

    print("Done")


if __name__ == "__main__":
    main()
