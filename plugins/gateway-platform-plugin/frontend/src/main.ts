import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router'

const style = document.createElement('style')
style.textContent = `
  :root {
    --gp-bg: #0b1020;
    --gp-bg-soft: #121a31;
    --gp-panel: rgba(14, 23, 45, 0.78);
    --gp-panel-strong: rgba(19, 32, 62, 0.92);
    --gp-border: rgba(166, 185, 255, 0.16);
    --gp-text: #e8eeff;
    --gp-text-soft: #9eaad1;
    --gp-accent: #7c9bff;
    --gp-accent-2: #5eead4;
    --gp-danger: #fb7185;
    --gp-shadow: 0 24px 80px rgba(4, 9, 24, 0.45);
  }

  body {
    margin: 0;
    min-height: 100vh;
    background:
      radial-gradient(circle at top left, rgba(124, 155, 255, 0.22), transparent 30%),
      radial-gradient(circle at 85% 12%, rgba(94, 234, 212, 0.18), transparent 26%),
      linear-gradient(180deg, #08101f 0%, #0b1020 48%, #11192d 100%);
    color: var(--gp-text);
    font-family: "SF Pro Display", "PingFang SC", "Helvetica Neue", sans-serif;
  }

  #app {
    min-height: 100vh;
  }

  .gp-shell {
    min-height: 100vh;
  }

  .gp-sidebar {
    background: linear-gradient(180deg, rgba(9, 15, 31, 0.94), rgba(11, 18, 35, 0.84));
    border-right: 1px solid var(--gp-border);
    backdrop-filter: blur(18px);
    box-shadow: inset -1px 0 0 rgba(255,255,255,0.04);
  }

  .gp-brand {
    padding: 28px 24px 18px;
  }

  .gp-brand-kicker {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 6px 10px;
    border: 1px solid rgba(124, 155, 255, 0.26);
    border-radius: 999px;
    color: var(--gp-accent-2);
    font-size: 12px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    background: rgba(124, 155, 255, 0.08);
  }

  .gp-brand-title {
    margin-top: 16px;
    font-size: 24px;
    font-weight: 700;
    letter-spacing: -0.04em;
    color: white;
  }

  .gp-brand-subtitle {
    margin-top: 8px;
    font-size: 13px;
    line-height: 1.7;
    color: var(--gp-text-soft);
  }

  .gp-main {
    background: transparent;
  }

  .gp-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 28px;
    border-bottom: 1px solid rgba(166, 185, 255, 0.08);
    background: rgba(9, 14, 28, 0.36);
    backdrop-filter: blur(16px);
  }

  .gp-header-title {
    font-size: 18px;
    font-weight: 600;
    color: white;
  }

  .gp-header-meta {
    font-size: 13px;
    color: var(--gp-text-soft);
  }

  .gp-content {
    padding: 28px;
  }

  .gp-panel-card {
    border: 1px solid var(--gp-border) !important;
    background: var(--gp-panel) !important;
    box-shadow: var(--gp-shadow) !important;
    border-radius: 22px !important;
    overflow: hidden;
    backdrop-filter: blur(24px);
  }

  .gp-panel-card .el-card__header {
    border-bottom: 1px solid rgba(255,255,255,0.06);
    background: linear-gradient(180deg, rgba(255,255,255,0.02), rgba(255,255,255,0.01));
    color: white;
  }

  .gp-drawer-modal {
    backdrop-filter: blur(8px);
    background: rgba(4, 9, 24, 0.46) !important;
  }

  .gp-drawer-panel {
    --el-bg-color: rgba(10, 16, 30, 0.96);
    --el-fill-color-blank: rgba(10, 16, 30, 0.96);
    --el-drawer-bg-color: rgba(10, 16, 30, 0.96);
    --el-text-color-primary: var(--gp-text);
    --el-text-color-regular: var(--gp-text-soft);
    --el-border-color-lighter: rgba(255,255,255,0.08);
    --el-border-color-light: rgba(255,255,255,0.08);
    --el-mask-color: rgba(4, 9, 24, 0.46);
    background: linear-gradient(180deg, rgba(8, 14, 28, 0.96), rgba(16, 24, 45, 0.92)) !important;
    border-left: 1px solid rgba(124, 155, 255, 0.16);
    box-shadow: -20px 0 60px rgba(0, 0, 0, 0.34);
    backdrop-filter: blur(22px);
  }

  .gp-drawer-panel,
  .gp-drawer-header-wrap,
  .gp-drawer-body-wrap,
  .gp-drawer-footer-wrap {
    background: transparent !important;
    color: var(--gp-text) !important;
  }

  .gp-drawer-header-wrap {
    margin-bottom: 0 !important;
    padding: 28px 28px 16px !important;
    border-bottom: 1px solid rgba(255,255,255,0.06);
  }

  .gp-drawer-body-wrap {
    padding: 22px 28px 18px !important;
  }

  .gp-drawer-footer-wrap {
    padding: 18px 28px 24px !important;
    border-top: 1px solid rgba(255,255,255,0.06);
  }

  .gp-drawer-panel .el-form-item__label,
  .gp-drawer-panel .el-checkbox,
  .gp-drawer-panel .el-switch__label,
  .gp-drawer-panel .el-radio__label {
    color: var(--gp-text-soft) !important;
  }

  .gp-drawer-panel .el-input__wrapper,
  .gp-drawer-panel .el-textarea__inner,
  .gp-drawer-panel .el-select__wrapper,
  .gp-drawer-panel .el-input-number,
  .gp-drawer-panel .el-input-number .el-input__wrapper {
    background: rgba(255,255,255,0.04) !important;
    border-color: rgba(255,255,255,0.08) !important;
    box-shadow: inset 0 0 0 1px rgba(255,255,255,0.04) !important;
    color: var(--gp-text) !important;
  }

  .gp-drawer-panel .el-input__inner,
  .gp-drawer-panel .el-textarea__inner,
  .gp-drawer-panel .el-select__selected-item,
  .gp-drawer-panel .el-input-number .el-input__inner {
    color: var(--gp-text) !important;
  }

  .gp-drawer-panel .el-card {
    background: rgba(255,255,255,0.04) !important;
    border-color: rgba(255,255,255,0.08) !important;
  }

  .gp-drawer-header {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .gp-drawer-title {
    font-size: 22px;
    font-weight: 700;
    letter-spacing: -0.04em;
    color: white;
  }

  .gp-drawer-subtitle {
    font-size: 13px;
    line-height: 1.7;
    color: var(--gp-text-soft);
  }

  .gp-drawer-body {
    padding-right: 4px;
  }

  .gp-drawer-footer {
    border-top: 1px solid rgba(255,255,255,0.06);
    padding-top: 14px;
  }

  .gp-table {
    --el-table-bg-color: transparent;
    --el-table-tr-bg-color: transparent;
    --el-table-header-bg-color: rgba(255,255,255,0.03);
    --el-table-border-color: rgba(255,255,255,0.06);
    --el-table-row-hover-bg-color: rgba(124, 155, 255, 0.08);
    --el-table-text-color: var(--gp-text);
    --el-table-header-text-color: var(--gp-text-soft);
  }

  .gp-table .cell {
    color: var(--gp-text);
  }

  .gp-muted {
    color: var(--gp-text-soft);
  }

  .gp-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 18px;
    margin-bottom: 22px;
  }

  .gp-stat {
    padding: 20px;
    border-radius: 20px;
    background: linear-gradient(145deg, rgba(14, 22, 42, 0.86), rgba(9, 17, 31, 0.78));
    border: 1px solid rgba(124,155,255,0.14);
    box-shadow: 0 16px 40px rgba(0,0,0,0.24);
  }

  .gp-stat-label {
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--gp-text-soft);
  }

  .gp-stat-value {
    margin-top: 12px;
    font-size: 32px;
    font-weight: 700;
    letter-spacing: -0.04em;
    color: white;
  }

  .gp-action-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
  }

  .gp-inline-actions {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    flex-wrap: nowrap;
    white-space: nowrap;
  }

  .gp-inline-actions .el-button {
    margin: 0 !important;
    min-width: 88px;
  }

  .gp-soft-button.el-button {
    border-radius: 14px;
    padding-inline: 18px;
    background: linear-gradient(135deg, rgba(33, 47, 84, 0.92), rgba(21, 34, 64, 0.82));
    border: 1px solid rgba(124, 155, 255, 0.18);
    color: #dce6ff;
    box-shadow: inset 0 1px 0 rgba(255,255,255,0.05), 0 10px 24px rgba(0,0,0,0.18);
  }

  .gp-soft-button.el-button:hover {
    color: white;
    border-color: rgba(124, 155, 255, 0.34);
    background: linear-gradient(135deg, rgba(43, 60, 104, 0.96), rgba(26, 40, 74, 0.9));
  }

  .gp-primary-button.el-button {
    border-radius: 14px;
    padding-inline: 18px;
    background: linear-gradient(135deg, #8aa4ff 0%, #5eead4 100%);
    border: none;
    color: #08101f;
    font-weight: 700;
    box-shadow: 0 14px 32px rgba(66, 119, 255, 0.28);
  }

  .gp-primary-button.el-button:hover {
    filter: brightness(1.03);
    box-shadow: 0 18px 38px rgba(66, 119, 255, 0.34);
  }

  .gp-drawer-panel .el-input-number {
    width: 100%;
  }

  .gp-drawer-panel .el-input-number__increase,
  .gp-drawer-panel .el-input-number__decrease {
    background: rgba(255,255,255,0.06) !important;
    color: #dce6ff !important;
    border: none !important;
    width: 34px;
  }

  .gp-drawer-panel .el-input-number__increase:hover,
  .gp-drawer-panel .el-input-number__decrease:hover {
    background: rgba(124, 155, 255, 0.18) !important;
    color: white !important;
  }

  .el-menu {
    border-right: none !important;
    background: transparent !important;
  }

  .el-menu-item {
    margin: 6px 14px;
    border-radius: 14px;
    color: var(--gp-text-soft) !important;
  }

  .el-menu-item.is-active {
    color: white !important;
    background: linear-gradient(135deg, rgba(124, 155, 255, 0.22), rgba(94, 234, 212, 0.16)) !important;
  }

  .el-empty {
    --el-empty-fill-color-0: rgba(124, 155, 255, 0.06);
    --el-empty-fill-color-1: rgba(124, 155, 255, 0.12);
    --el-empty-fill-color-2: rgba(124, 155, 255, 0.18);
    --el-empty-fill-color-3: rgba(94, 234, 212, 0.14);
    --el-empty-fill-color-4: rgba(94, 234, 212, 0.2);
    --el-empty-fill-color-5: rgba(255, 255, 255, 0.08);
    --el-empty-fill-color-6: rgba(255, 255, 255, 0.12);
    --el-empty-fill-color-7: rgba(124, 155, 255, 0.22);
    --el-empty-fill-color-8: rgba(94, 234, 212, 0.24);
    --el-empty-fill-color-9: rgba(255, 255, 255, 0.16);
  }

  .el-empty__description p {
    color: var(--gp-text-soft);
  }

  .el-empty__image svg {
    filter: drop-shadow(0 18px 36px rgba(0, 0, 0, 0.28));
  }

  .el-overlay-message-box {
    backdrop-filter: blur(8px);
    background: rgba(4, 9, 24, 0.46) !important;
  }

  .el-message-box {
    background: linear-gradient(180deg, rgba(8, 14, 28, 0.97), rgba(16, 24, 45, 0.94)) !important;
    border: 1px solid rgba(124, 155, 255, 0.16) !important;
    box-shadow: 0 26px 80px rgba(0, 0, 0, 0.42) !important;
    border-radius: 22px !important;
    backdrop-filter: blur(20px);
  }

  .el-message-box__title,
  .el-message-box__message,
  .el-message-box__message p,
  .el-message-box__status {
    color: var(--gp-text) !important;
  }

  .el-message-box__header {
    padding: 24px 24px 10px !important;
    border-bottom: 1px solid rgba(255,255,255,0.06);
  }

  .el-message-box__content {
    padding: 18px 24px 22px !important;
  }

  .el-message-box__btns {
    padding: 0 24px 22px !important;
    border-top: 1px solid rgba(255,255,255,0.06);
    margin-top: 6px;
  }

  .el-message-box__btns .el-button:first-child {
    border-radius: 14px;
    background: linear-gradient(135deg, rgba(33, 47, 84, 0.92), rgba(21, 34, 64, 0.82));
    border: 1px solid rgba(124, 155, 255, 0.18);
    color: #dce6ff;
    box-shadow: inset 0 1px 0 rgba(255,255,255,0.05), 0 10px 24px rgba(0,0,0,0.18);
  }

  .el-message-box__btns .el-button--primary {
    border-radius: 14px;
    background: linear-gradient(135deg, #8aa4ff 0%, #5eead4 100%);
    border: none;
    color: #08101f;
    font-weight: 700;
    box-shadow: 0 14px 32px rgba(66, 119, 255, 0.28);
  }

  .el-message {
    background: linear-gradient(135deg, rgba(14, 22, 42, 0.96), rgba(12, 20, 36, 0.94)) !important;
    border: 1px solid rgba(124, 155, 255, 0.16) !important;
    color: var(--gp-text) !important;
    box-shadow: 0 18px 40px rgba(0,0,0,0.28);
    backdrop-filter: blur(18px);
  }

  .el-message .el-message__content {
    color: var(--gp-text) !important;
  }
`
document.head.appendChild(style)

createApp(App).use(router).use(ElementPlus).mount('#app')
