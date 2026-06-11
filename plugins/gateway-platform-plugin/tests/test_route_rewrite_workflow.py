from pathlib import Path
import unittest


ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "frontend" / "src" / "components" / "RouteFormDrawer.vue"


class RouteRewriteWorkflowTest(unittest.TestCase):
    def read_source(self) -> str:
        return SOURCE.read_text()

    def test_new_route_save_keeps_drawer_open_for_rewrite_editing(self) -> None:
        source = self.read_source()

        self.assertIn("if (isEdit.value) close()", source)
        self.assertNotIn("emit('saved', route)\n    close()", source)


if __name__ == "__main__":
    unittest.main()
