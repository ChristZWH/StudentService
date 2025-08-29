-- 定义ID范围
min_id = 1
max_id = 500

-- 初始化函数
function init(args)
    math.randomseed(os.time())
end

-- 请求生成函数
request = function()
    -- 生成随机ID (1-500)
    local id_num = math.random(min_id, max_id)
    
    -- 格式化为6位数字字符串
    local id_str = string.format("%06d", id_num)
    
    -- 构建完整ID (STU000001格式)
    local full_id = "STU" .. id_str
    
    -- 创建请求路径
    local path = "/students/" .. full_id
    
    -- 返回HTTP GET请求
    return wrk.format("GET", path)
end