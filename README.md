# isb

Integrated Service Back-office.

# generate mock

```
mockgen -source=E:\code\golang\isb\src\service\captcha.go   -destination=E:\code\golang\isb\src\service\mocks\captcha.mock.gen.go -package=svcmock
mockgen -source=E:\code\golang\isb\src\service\user.go   -destination=E:\code\golang\isb\src\service\mocks\user.mock.gen.go -package=svcmock
mockgen -source=E:\code\golang\isb\src\repository\user.go   -destination=E:\code\golang\isb\src\repository\mocks\user.mock.gen.go -package=repomock
mockgen -source=E:\code\golang\isb\src\repository\dao\user.go   -destination=E:\code\golang\isb\src\repository\dao\mocks\user.mock.gen.go -package=daomock
mockgen -source=E:\code\golang\isb\src\repository\cache\user.go   -destination=E:\code\golang\isb\src\repository\cache\mocks\user.mock.gen.go -package=cachemock
mockgen -package=redismocks -destination=E:\code\golang\isb\src\repository\cache\redismocks\cmdable.mock.gen.go github.com/redis/go-redis/v9 Cmdable
mockgen -source=E:\code\golang\isb\src\pkg\ratelimit\ratelimit.go   -destination=E:\code\golang\isb\src\pkg\ratelimit\mocks\limiter.mock.gen.go -package=limitermock
mockgen -source=E:\code\golang\isb\src\service\sms\types.go   -destination=E:\code\golang\isb\src\service\sms\mocks\sms.mock.gen.go -package=smsmock
```
