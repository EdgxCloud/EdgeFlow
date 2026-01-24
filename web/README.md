# EdgeFlow Web - رابط کاربری

رابط کاربری EdgeFlow ساخته شده با React, TypeScript, و TailwindCSS

## 🚀 راه‌اندازی

### پیش‌نیازها
- Node.js 18+
- npm یا yarn

### نصب

```bash
# نصب وابستگی‌ها
npm install

# اجرا در حالت توسعه
npm run dev

# Build برای production
npm run build

# Preview build
npm run preview
```

## 📁 ساختار پروژه

```
web/
├── src/
│   ├── components/       # کامپوننت‌های قابل استفاده مجدد
│   │   ├── Layout.tsx
│   │   ├── Header.tsx
│   │   └── Sidebar.tsx
│   ├── pages/            # صفحات اصلی
│   │   ├── Dashboard.tsx
│   │   ├── Workflows.tsx
│   │   ├── Editor.tsx
│   │   ├── Executions.tsx
│   │   └── Settings.tsx
│   ├── stores/           # State management با Zustand
│   │   └── flowStore.ts
│   ├── lib/              # توابع کمکی و API client
│   │   └── api.ts
│   ├── App.tsx           # کامپوننت اصلی
│   ├── main.tsx          # نقطه ورود
│   └── index.css         # استایل‌های global
├── index.html
├── package.json
├── vite.config.ts
├── tailwind.config.js
└── tsconfig.json
```

## 🎨 تکنولوژی‌ها

- **React 18** - UI library
- **TypeScript** - Type safety
- **Vite** - Build tool
- **TailwindCSS** - Utility-first CSS
- **React Router** - Routing
- **Zustand** - State management
- **Axios** - HTTP client
- **Lucide React** - Icons
- **@xyflow/react** - Flow editor (آماده برای نصب)

## 🌐 API Integration

Frontend با backend از طریق REST API در `/api/v1` ارتباط برقرار می‌کند.

### تنظیمات Proxy

در `vite.config.ts`:
```typescript
server: {
  proxy: {
    '/api': 'http://localhost:8080',
    '/ws': 'ws://localhost:8080',
  }
}
```

## 🎯 ویژگی‌های پیاده‌سازی شده

- ✅ ساختار پروژه و تنظیمات
- ✅ Layout اصلی با Sidebar و Header
- ✅ داشبورد با آمار
- ✅ صفحه لیست فلوها
- ✅ State management با Zustand
- ✅ API client کامل
- ✅ Dark mode support
- ✅ رابط کاربری فارسی (RTL)
- 🔄 Flow Editor - در حال توسعه
- 🔄 Real-time updates با WebSocket - در حال توسعه

## 📝 TODO

- [ ] پیاده‌سازی کامل Flow Editor با React Flow
- [ ] WebSocket برای real-time updates
- [ ] صفحه Executions
- [ ] صفحه Settings
- [ ] Authentication
- [ ] Form validation
- [ ] Error boundaries
- [ ] Loading states
- [ ] Toast notifications
- [ ] Unit tests

## 🔧 Development

### Run dev server
```bash
npm run dev
# سرور روی http://localhost:3000 اجرا می‌شود
```

### Build
```bash
npm run build
# فایل‌های build در پوشه dist/ ایجاد می‌شوند
```

### Lint
```bash
npm run lint
```

## 🌍 محیط‌های متغیر

ایجاد فایل `.env.local`:

```env
VITE_API_URL=http://localhost:8080/api/v1
```

## 📱 Responsive Design

رابط کاربری برای تمام سایزهای صفحه بهینه شده است:
- موبایل (< 768px)
- تبلت (768px - 1024px)
- دسکتاپ (> 1024px)

## 🎨 طراحی

- فارسی و RTL
- Dark mode
- رنگ‌بندی سازگار
- انیمیشن‌های smooth
- دسترسی‌پذیری (a11y)

---

**نسخه:** 0.1.0
**وضعیت:** در حال توسعه
