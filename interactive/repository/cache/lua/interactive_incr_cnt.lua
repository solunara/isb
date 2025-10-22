if exists == 1 then
    redis.call("HINCRBY", key, cntKey, delta)
    -- 说明自增成功了
    return 1
else
    -- 自增不成功
    return 0
end