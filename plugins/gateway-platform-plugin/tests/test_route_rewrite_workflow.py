from pathlib import Path
import unittest


ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "frontend" / "src" / "components" / "RouteFormDrawer.vue"
STORAGE_STATUS_SOURCE = ROOT / "frontend" / "src" / "components" / "DataStorageStatus.vue"
DASHBOARD_SOURCE = ROOT / "frontend" / "src" / "pages" / "DashboardPage.vue"
ROUTES_SOURCE = ROOT / "frontend" / "src" / "pages" / "RoutesPage.vue"


class RouteRewriteWorkflowTest(unittest.TestCase):
    def read_source(self) -> str:
        return SOURCE.read_text()

    def test_new_route_save_keeps_drawer_open_for_rewrite_editing(self) -> None:
        source = self.read_source()

        self.assertIn("const wasEdit = isEdit.value", source)
        self.assertIn("if (wasEdit) close()", source)
        self.assertNotIn("emit('saved', route)\n    close()", source)

    def test_route_save_displays_backend_error(self) -> None:
        source = self.read_source()

        self.assertIn("apiErrorMessage(error, '未知错误')", source)
        self.assertIn("保存 Route 路由失败：", source)

    def test_dashboard_create_action_opens_route_drawer(self) -> None:
        dashboard = DASHBOARD_SOURCE.read_text()
        routes = ROUTES_SOURCE.read_text()

        self.assertIn("query: { create: '1' }", dashboard)
        self.assertIn("route.query.create === '1'", routes)
        self.assertIn("handleCreate()", routes)

    def test_dashboard_displays_persistent_data_directory(self) -> None:
        dashboard = DASHBOARD_SOURCE.read_text()
        storage_status = STORAGE_STATUS_SOURCE.read_text()

        self.assertIn("<DataStorageStatus />", dashboard)
        self.assertIn("Data Directory 数据目录", storage_status)
        self.assertIn("database_writable", storage_status)
        self.assertIn("READ / WRITE", storage_status)


if __name__ == "__main__":
    unittest.main()
