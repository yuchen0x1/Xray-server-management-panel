import importlib.util
import unittest
from pathlib import Path


SPEC = importlib.util.spec_from_file_location(
    "smoke_archive", Path(__file__).with_name("smoke-archive.py"))
SMOKE_ARCHIVE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(SMOKE_ARCHIVE)


class ExpectedPanelVersionTests(unittest.TestCase):
    def test_local_archive_uses_embedded_version(self):
        self.assertEqual(SMOKE_ARCHIVE.expected_panel_version("3.13.2", {}), "3.13.2")

    def test_tag_archive_uses_embedded_version(self):
        environment = {"GITHUB_REF": "refs/tags/v3.13.2", "GITHUB_SHA": "a" * 40}
        self.assertEqual(SMOKE_ARCHIVE.expected_panel_version("3.13.2", environment), "3.13.2")

    def test_non_tag_archives_use_their_exact_commit(self):
        for ref in ("refs/heads/main", "refs/pull/26/merge"):
            with self.subTest(ref=ref):
                environment = {"GITHUB_REF": ref, "GITHUB_SHA": "27137140" + "a" * 32}
                self.assertEqual(SMOKE_ARCHIVE.expected_panel_version("3.13.2", environment),
                                 "dev+27137140")

    def test_non_tag_archive_rejects_missing_or_invalid_commit(self):
        for commit in ("", "27137140", "g" * 40, "a" * 39, "a" * 41):
            with self.subTest(commit=commit):
                with self.assertRaisesRegex(RuntimeError, "requires a valid GITHUB_SHA"):
                    SMOKE_ARCHIVE.expected_panel_version(
                        "3.13.2", {"GITHUB_REF": "refs/heads/main", "GITHUB_SHA": commit})


if __name__ == "__main__":
    unittest.main()
