--获取验证码在redis中的key, phone_code:login:139xxxxxxxx
local key = KEYS[1]

--获取验证码的验证次数, phone_code:login:139xxxxxxxx:cnt
local cntkey = key..":cnt"

--获取验证码
local val = ARGV[1]

--获取验证码过期时间
local ttl = tonumber(redis.call("ttl",key))

--key存在但是没有过期时间
if ttl == -1 then
    --系统错误或人为操作，没有设置过期时间
    return -2
--过期时间小于9分钟，说明已经过了至少1分钟，可以发
elseif ttl == -2 or ttl < 540 then
    redis.call("set", key, val)
    redis.call("expire", key, 600)
    redis.call("set", cntkey, 3)
    redis.call("expire", cntkey, 600)
    return 0
else
    --发送太频繁
    return -1
end