package handleers

import (
	"StudentService/database"
	"StudentService/models"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 1. 创建学生
func CreateStudent(c *gin.Context) {
	var student models.Student
	if err := c.ShouldBindJSON(&student); err != nil {
		log.Printf("创建学生请求解析失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	// // 执行数据库操作
	// sqlstr := "INSERT INTO students (id, name, tel, study) VALUES (?, ?, ?, ?)"
	// result, err := database.DB.Exec(sqlstr, student.ID, student.Name, student.Tel, student.Study)
	// if err != nil {
	// 	log.Printf("创建学生失败: %v, 数据: %+v", err, student)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "创建学生失败"})
	// 	return
	// }
	// // 获取影响行数
	// rowsAffected, _ := result.RowsAffected()
	// log.Printf("成功创建学生 %s, 影响行数: %d", student.ID, rowsAffected)

	// --- 使用 GORM 创建记录 ---
	result := database.GormDB.Create(&student) // 传入结构体指针
	if result.Error != nil {
		log.Printf("创建学生失败: %v, 数据: %+v", result.Error, student)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建学生失败"})
		return
	}
	log.Printf("成功创建学生 %s, ID: %v", student.ID, student.ID) // GORM 会填充主键
	// --- GORM 创建结束 ---

	// 清除学生列表缓存！！！
	ctx := context.Background()
	// func (c redis.cmdable) Del(ctx context.Context, keys ...string) *redis.IntCmd
	if err := database.RedisClient.Del(ctx, "students:list").Err(); err != nil {
		log.Printf("清除列表缓存失败: %v", err)
	}

	c.JSON(http.StatusCreated, student)
}

// 2. 获取所有学生
func ListStudents(c *gin.Context) {
	ctx := context.Background()
	cacheKey := "students:list"

	// 从Redis获取列表缓存
	// func (c redis.cmdable) Get(ctx context.Context, key string) *redis.StringCmd
	cachedData, err := database.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var students []models.Student
		if err := json.Unmarshal([]byte(cachedData), &students); err == nil {
			log.Printf("从缓存获取学生列表 (数量: %d)", len(students))
			c.JSON(http.StatusOK, students)
			return
		}
		log.Printf("列表缓存数据解析失败: %v", err)
	}

	// // 没获取到数据，继续在数据库中查询
	// sqlstr := "SELECT id, name, tel, study FROM students"
	// rows, err := database.DB.Query(sqlstr)
	// if err != nil {
	// 	log.Printf("查询学生列表失败: %v", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "获取学生列表失败"})
	// 	return
	// }
	// defer rows.Close()

	// var students []models.Student
	// for rows.Next() {
	// 	var temp models.Student
	// 	if err := rows.Scan(&temp.ID, &temp.Name, &temp.Tel, &temp.Study); err != nil {
	// 		log.Printf("解析学生数据失败: %v", err)
	// 		continue
	// 	}
	// 	students = append(students, temp)
	// }
	// // 检查迭代错误是一个良好的习惯   嘻嘻
	// if err := rows.Err(); err != nil {
	// 	log.Printf("遍历学生数据出错: %v", err)
	// }

	// --- 使用 GORM 查询所有记录 ---
	var students []models.Student
	result := database.GormDB.Find(&students) // Find 查询所有记录
	if result.Error != nil {
		log.Printf("查询学生列表失败: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取学生列表失败"})
		return
	}
	// --- GORM 查询结束 ---

	log.Printf("从数据库获取学生列表 (数量: %d)", len(students))

	// 存入缓存 (设置2分钟过期)
	studentsJSON, _ := json.Marshal(students)
	// func (c redis.cmdable) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	if err := database.RedisClient.Set(ctx, cacheKey, studentsJSON, 2*time.Minute).Err(); err != nil {
		log.Printf("缓存学生列表失败: %v", err)
	}

	c.JSON(http.StatusOK, students)
}

// 3. 获取单个学生
func GetStudent(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	cacheKey := "student:" + id

	// 从Redis获取列表缓存
	cachedData, err := database.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var student models.Student
		if err := json.Unmarshal([]byte(cachedData), &student); err == nil {
			log.Printf("从缓存获取学生 %s", id)
			c.JSON(http.StatusOK, student)
			return
		}
		log.Printf("缓存数据解析失败: %v", err)
	}

	// // 没获取到数据，继续在数据库中查询
	// sqlstr := "SELECT id, name, tel, study FROM students WHERE id = ?"
	// var student models.Student
	// err = database.DB.QueryRow(sqlstr, id).Scan(&student.ID, &student.Name, &student.Tel, &student.Study)

	// switch {
	// case err == sql.ErrNoRows:
	// 	log.Printf("学生不存在: %s", id)
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
	// 	return
	// case err != nil:
	// 	log.Printf("查询学生 %s 失败: %v", id, err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "获取学生信息失败"})
	// 	return
	// }

	// --- 使用 GORM 根据主键查询 ---
	var student models.Student
	result := database.GormDB.First(&student, "id = ?", id) // First 查找第一个匹配的记录
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {    // 需要导入 "errors" 和 "gorm.io/gorm"
		log.Printf("学生不存在: %s", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
		return
	}
	if result.Error != nil {
		log.Printf("查询学生 %s 失败: %v", id, result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取学生信息失败"})
		return
	}
	// --- GORM 查询结束 ---

	log.Printf("从数据库获取学生 %s", id)

	// 存入缓存
	studentJSON, _ := json.Marshal(student)
	if err := database.RedisClient.Set(ctx, cacheKey, studentJSON, 5*time.Minute).Err(); err != nil {
		log.Printf("缓存学生数据失败: %v", err)
	}

	c.JSON(http.StatusOK, student)
}

// 4. 更新学生 (更新后清除缓存)
func UpdateStudent(c *gin.Context) {
	id := c.Param("id")
	var updateData struct {
		Name  string `json:"name"`
		Tel   string `json:"tel"`
		Study string `json:"study"`
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		log.Printf("更新请求解析失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	// sqlstr := "UPDATE students SET name = ?, tel = ?, study = ? WHERE id = ?"
	// result, err := database.DB.Exec(sqlstr, updateData.Name, updateData.Tel, updateData.Study, id)
	// if err != nil {
	// 	log.Printf("更新学生 %s 失败: %v", id, err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "更新学生失败"})
	// 	return
	// }

	// rowsAffected, err := result.RowsAffected()
	// if err != nil {
	// 	log.Printf("获取影响行数失败: %v", err)
	// }
	// if rowsAffected == 0 {
	// 	log.Printf("更新失败: 学生不存在 %s", id)
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
	// 	return
	// }
	// log.Printf("成功更新学生 %s, 影响行数: %d", id, rowsAffected)

	// --- 使用 GORM 更新记录 ---
	// 方法 1: 先查询再更新 (可以更精确地知道是否找到记录)
	var student models.Student
	result := database.GormDB.First(&student, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Printf("更新失败: 学生不存在 %s", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
		return
	}
	if result.Error != nil {
		log.Printf("查询学生 %s 失败: %v", id, result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新学生失败"})
		return
	}

	// 更新字段 (仅更新非空字段或有变化的字段)
	updates := map[string]interface{}{}
	if updateData.Name != "" {
		updates["name"] = updateData.Name
	}
	if updateData.Tel != "" {
		updates["tel"] = updateData.Tel
	}
	if updateData.Study != "" {
		updates["study"] = updateData.Study
	}

	if len(updates) > 0 {
		result = database.GormDB.Model(&student).Updates(updates) // 更新指定字段
		if result.Error != nil {
			log.Printf("更新学生 %s 失败: %v", id, result.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新学生失败"})
			return
		}
		log.Printf("成功更新学生 %s", id)
	} else {
		log.Printf("没有需要更新的字段 for 学生 %s", id)
	}
	// --- GORM 更新结束 ---

	// 清除该学生的缓存
	ctx := context.Background()
	cacheKey := "student:" + id
	if err := database.RedisClient.Del(ctx, cacheKey).Err(); err != nil {
		log.Printf("清除缓存失败: %v", err)
	} else {
		log.Printf("已清除学生缓存: %s", cacheKey)
	}
	// 清除学生列表缓存
	if err := database.RedisClient.Del(ctx, "students:list").Err(); err != nil {
		log.Printf("清除列表缓存失败: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "更新成功"})
}

// 5. 删除学生 (清除缓存)
func DeleteStudent(c *gin.Context) {
	id := c.Param("id")
	// sqlstr := "DELETE FROM students WHERE id = ?"
	// result, err := database.DB.Exec(sqlstr, id)
	// if err != nil {
	// 	log.Printf("删除学生 %s 失败: %v", id, err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "删除学生失败"})
	// 	return
	// }

	// rowsAffected, err := result.RowsAffected()
	// if err != nil {
	// 	log.Printf("获取影响行数失败: %v", err)
	// }

	// if rowsAffected == 0 {
	// 	log.Printf("删除失败: 学生不存在 %s", id)
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
	// 	return
	// }
	// log.Printf("成功删除学生 %s, 影响行数: %d", id, rowsAffected)

	// --- 使用 GORM 删除记录 ---
	// 方法 1: 先查询再删除 (可以更精确地知道是否找到记录)
	var student models.Student
	result := database.GormDB.First(&student, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Printf("删除失败: 学生不存在 %s", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
		return
	}
	if result.Error != nil {
		log.Printf("查询学生 %s 失败: %v", id, result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除学生失败"})
		return
	}

	result = database.GormDB.Delete(&student) // 删除记录
	if result.Error != nil {
		log.Printf("删除学生 %s 失败: %v", id, result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除学生失败"})
		return
	}
	log.Printf("成功删除学生 %s", id)

	// 清除该学生的缓存
	ctx := context.Background()
	cacheKey := "student:" + id
	if err := database.RedisClient.Del(ctx, cacheKey).Err(); err != nil {
		log.Printf("清除缓存失败: %v", err)
	} else {
		log.Printf("已清除学生缓存: %s", cacheKey)
	}

	// 清除学生列表缓存
	if err := database.RedisClient.Del(ctx, "students:list").Err(); err != nil {
		log.Printf("清除列表缓存失败: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "删除成功"})
}
