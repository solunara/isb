local key = KEYS[1]
local cntkey = key.."cnt"

--获取用户输入的验证码
local expectedCaptcha = ARGV[1]

--获取redis中存储的验证码
local code = redis.call("get",key)

--用户输错了,直接返回
if expectedCaptcha ~= code then
    redis.call("decr",cntkey)
    return -2
end

--获取验证码的验证次数
local cnt = tonumber(redis.call("get",cntkey))

if cnt <= 0 then
    --用户一直输错
    return -1
else
    redis.call("set",cntkey,-1)
    return 0
end