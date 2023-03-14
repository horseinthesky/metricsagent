# 📏 metricsagent

`metricsagent` is a main project of Advanced Go developer course from Yandex Praticum.

## ✨ Features

- 🚀 Client side sends runtime metrics to server side. Both HTTP and gRPC supported
- 🔒 Metric values are encrypted with [GCM](https://en.wikipedia.org/wiki/Galois/Counter_Mode)
- 💪 Async execution for improved performance

## 📊 AutoTests

Project autotests are available here:
https://github.com/Yandex-Practicum/go-autotests/tree/main/cmd/gophermarttest

### Updates

To be able to get updates for the test suite run:

```
git remote add -m master template https://github.com/yandex-praktikum/go-musthave-diploma-tpl.git
```

To update the test suite source code run:

```
git fetch template && git checkout template/master .github
```

Then add changes to your repo.
