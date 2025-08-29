request = function()
    if math.random() < 0.1 then
        return wrk.format("GET", "/students")
    else
        -- 生成 1-500 的随机 ID（均匀分布替代 zipf）
        local id = math.random(1, 500)
        return wrk.format("GET", "/students/STU" .. string.format("%06d", id))
    end
end