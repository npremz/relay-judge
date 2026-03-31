import contextlib
import importlib.util
import io
import json
import pathlib
import sys
import time
import traceback


def emit(payload):
    sys.stdout.write(json.dumps(payload))
    sys.stdout.flush()


def capture(callable_):
    stdout_buffer = io.StringIO()
    stderr_buffer = io.StringIO()
    with contextlib.redirect_stdout(stdout_buffer), contextlib.redirect_stderr(stderr_buffer):
        value = callable_()
    return value, stdout_buffer.getvalue(), stderr_buffer.getvalue()


def safe_value(value):
    try:
        return json.loads(json.dumps(value))
    except TypeError as exc:
        raise TypeError(f"return value is not JSON serializable: {exc}") from exc


def load_submission(submission_path):
    module_name = f"relay_submission_{int(time.time() * 1000000)}"
    spec = importlib.util.spec_from_file_location(module_name, submission_path)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"unable to load module from {submission_path}")

    module = importlib.util.module_from_spec(spec)
    _, captured_stdout, captured_stderr = capture(lambda: spec.loader.exec_module(module))
    return module, captured_stdout, captured_stderr


def main():
    if len(sys.argv) != 2:
        emit({"status": "load_error", "error": "wrapper expects a submission path"})
        return 0

    submission_path = pathlib.Path(sys.argv[1]).resolve()
    if not submission_path.exists():
        emit({"status": "load_error", "error": f"submission file not found: {submission_path}"})
        return 0

    try:
        payload = json.load(sys.stdin)
    except Exception:
        emit({"status": "load_error", "error": traceback.format_exc()})
        return 0

    try:
        module, captured_stdout, captured_stderr = load_submission(submission_path)
    except Exception:
        emit({"status": "load_error", "error": traceback.format_exc()})
        return 0

    function_name = payload.get("function_name")
    function = getattr(module, function_name, None)
    if not callable(function):
        message = f"missing callable {function_name!r}"
        if captured_stdout or captured_stderr:
            message = f"{message}\nstdout:\n{captured_stdout}\nstderr:\n{captured_stderr}"
        emit({"status": "load_error", "error": message})
        return 0

    tests = []
    for test in payload.get("tests", []):
        try:
            start = time.perf_counter()

            def call():
                return function(*test.get("args", []))

            actual, test_stdout, test_stderr = capture(call)
            duration_ms = (time.perf_counter() - start) * 1000

            tests.append(
                {
                    "name": test.get("name"),
                    "group": test.get("group"),
                    "status": "ok",
                    "actual": safe_value(actual),
                    "duration_ms": duration_ms,
                    "stdout": test_stdout,
                    "stderr": test_stderr,
                }
            )
        except Exception:
            tests.append(
                {
                    "name": test.get("name"),
                    "group": test.get("group"),
                    "status": "runtime_error",
                    "error": traceback.format_exc(),
                }
            )
            emit({"status": "ok", "tests": tests})
            return 0

    emit({"status": "ok", "tests": tests})
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
