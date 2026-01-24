# EdgeFlow - راهنمای شروع سریع

## نصب و راه‌اندازی

### پیش‌نیازها

- Go 1.21 یا بالاتر
- (اختیاری) Node.js 18+ برای توسعه فرانت‌اند
- (اختیاری) Docker برای استقرار

### نصب از سورس

```bash
# کلون کردن مخزن
git clone https://github.com/yourusername/edgeflow.git
cd edgeflow

# دانلود وابستگی‌ها
go mod download

# ساخت برنامه
make build

# اجرای برنامه
./bin/edgeflow
```

### اجرا با Docker

```bash
# ساخت تصویر
docker build -t edgeflow:latest .

# اجرا
docker run -p 8080:8080 edgeflow:latest
```

### اجرا با Docker Compose

```bash
docker-compose up -d
```

## استفاده اولیه

### 1. دسترسی به رابط کاربری

پس از اجرای برنامه، به آدرس زیر بروید:

```
http://localhost:8080
```

### 2. ایجاد اولین Flow

1. روی دکمه "Flow جدید" کلیک کنید
2. یک نام برای Flow خود انتخاب کنید
3. Node های مورد نظر را از پالت به کانوس بکشید
4. Node ها را به هم وصل کنید
5. روی دکمه "Deploy" کلیک کنید

### 3. مثال ساده - Hello World

این مثال یک پیام ساده را هر 5 ثانیه ارسال و نمایش می‌دهد:

1. یک **Inject Node** اضافه کنید:
   - Interval: 5 ثانیه
   - Payload: `{"message": "Hello World"}`

2. یک **Debug Node** اضافه کنید:
   - Output to: Console

3. Inject Node را به Debug Node وصل کنید

4. روی Deploy کلیک کنید و خروجی را در کنسول مشاهده کنید

## API Documentation

EdgeFlow یک REST API کامل ارائه می‌دهد:

### Health Check

```bash
curl http://localhost:8080/api/v1/health
```

### لیست Flowها

```bash
curl http://localhost:8080/api/v1/flows
```

### ایجاد Flow جدید

```bash
curl -X POST http://localhost:8080/api/v1/flows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Flow",
    "description": "A test flow"
  }'
```

### شروع یک Flow

```bash
curl -X POST http://localhost:8080/api/v1/flows/{flow_id}/start
```

## پیکربندی

فایل پیکربندی اصلی در `configs/default.yaml` قرار دارد.

برای تغییرات محلی، یک فایل `configs/local.yaml` ایجاد کنید که تنظیمات پیش‌فرض را override می‌کند.

### متغیرهای محیطی

```bash
# تنظیم پورت سرور
export EDGEFLOW_SERVER_PORT=3000

# تنظیم سطح لاگ
export EDGEFLOW_LOGGER_LEVEL=debug

# تنظیم مسیر دیتابیس
export EDGEFLOW_DATABASE_PATH=/data/edgeflow.db
```

## Node های موجود

### Core Nodes

- **Inject**: ارسال پیام در بازه‌های زمانی مشخص
- **Debug**: نمایش خروجی برای دیباگ

### (در دست توسعه)

- **HTTP Request**: ارسال درخواست HTTP
- **MQTT In/Out**: ارتباط با Broker های MQTT
- **Function**: اجرای کد JavaScript سفارشی
- **GPIO**: کنترل پین‌های GPIO در Raspberry Pi

## توسعه

برای شروع توسعه، [راهنمای مشارکت](../CONTRIBUTING.md) را مطالعه کنید.

### اجرا در حالت توسعه

```bash
# اجرا با hot reload
make dev

# اجرای تست‌ها
make test

# بررسی کیفیت کد
make lint
```

## مشکلات رایج

### پورت 8080 در حال استفاده است

پورت را تغییر دهید:

```bash
export EDGEFLOW_SERVER_PORT=3000
./bin/edgeflow
```

### خطای دسترسی به GPIO

در Raspberry Pi، ممکن است نیاز به اجرا با sudo باشید:

```bash
sudo ./bin/edgeflow
```

## منابع بیشتر

- [مستندات کامل](https://edgeflow.io/docs)
- [مثال‌های بیشتر](./examples/)
- [API Reference](./api-reference.md)
- [چک‌لیست توسعه](../DEVELOPMENT_CHECKLIST.md)

## پشتیبانی

- گزارش باگ: [GitHub Issues](https://github.com/yourusername/edgeflow/issues)
- بحث و گفتگو: [GitHub Discussions](https://github.com/yourusername/edgeflow/discussions)
- ایمیل: support@edgeflow.io
