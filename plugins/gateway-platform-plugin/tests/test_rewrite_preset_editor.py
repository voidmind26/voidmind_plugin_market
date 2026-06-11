from pathlib import Path
import unittest


ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "frontend" / "src" / "components" / "RewriteTableEditor.vue"


class RewritePresetEditorTest(unittest.TestCase):
    def read_source(self) -> str:
        return SOURCE.read_text()

    def test_custom_rewrite_preset_falls_back_to_custom_rule(self) -> None:
        source = self.read_source()

        self.assertIn("const customPresetKey = 'custom'", source)
        self.assertIn("item.__presetKey = preset?.key ?? customPresetKey", source)
        self.assertIn("return 'Custom Rule'", source)
        self.assertNotIn("preset?.key ?? presetGroups[type][0]?.key", source)

    def test_type_change_applies_default_preset_fields(self) -> None:
        source = self.read_source()

        self.assertIn("const handleTypeChange = (item: any)", source)
        self.assertIn("item.__presetKey = currentPresets(item)[0]?.key ?? customPresetKey", source)
        self.assertIn("applyPreset(item)", source)
        self.assertIn('@change="handleTypeChange(item)"', source)
        self.assertNotIn('@change="hydratePreset(item)"', source)

    def test_custom_preset_is_visible_in_select(self) -> None:
        source = self.read_source()

        self.assertIn('v-if="item.__presetKey === customPresetKey"', source)
        self.assertIn('label="Custom Rule"', source)
        self.assertIn(':value="customPresetKey"', source)

    def test_manual_target_or_template_edits_rehydrate_preset(self) -> None:
        source = self.read_source()

        self.assertIn("const handleRewriteFieldChange = (item: any)", source)
        self.assertIn('@input="handleRewriteFieldChange(item)"', source)


if __name__ == "__main__":
    unittest.main()
