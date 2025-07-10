package handlers

import (
	"StudentService/database"
	"StudentService/models"
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 创建学生
func CreateStudent(c *gin.Context) {
	var student models.Student
	if err := c.ShouldBindJSON(&student); err != nil {
		log.Printf("创建学生请求解析失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"}) //400
		return
	}

	// 执行数据库操作
	result, err := database.DB.Exec(
		"INSERT INTO students (id, name, tel, study) VALUES (?, ?, ?, ?)",
		student.ID, student.Name, student.Tel, student.Study,
	)

	if err != nil {
		log.Printf("创建学生失败: %v, 数据: %+v", err, student)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建学生失败"}) //500
		return
	}

	// 获取影响行数
	rowsAffected, _ := result.RowsAffected()
	log.Printf("成功创建学生 %s, 影响行数: %d", student.ID, rowsAffected)

	c.JSON(http.StatusCreated, student) //201
}

// 获取所有学生
func ListStudents(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, name, tel, study FROM students")
	if err != nil {
		log.Printf("查询学生列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取学生列表失败"})
		return
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(&s.ID, &s.Name, &s.Tel, &s.Study); err != nil {
			log.Printf("解析学生数据失败: %v", err)
			continue // 继续处理其他行
		}
		students = append(students, s)
	}

	if err := rows.Err(); err != nil {
		log.Printf("遍历学生数据出错: %v", err)
	}

	c.JSON(http.StatusOK, students)
}

// 获取单个学生
func GetStudent(c *gin.Context) {
	id := c.Param("id")

	var student models.Student
	err := database.DB.QueryRow(
		"SELECT id, name, tel, study FROM students WHERE id = ?", id,
	).Scan(&student.ID, &student.Name, &student.Tel, &student.Study)

	switch {
	case err == sql.ErrNoRows:
		c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
		return
	case err != nil:
		log.Printf("查询学生 %s 失败: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取学生信息失败"})
		return
	}

	c.JSON(http.StatusOK, student)
}

// 更新学生
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

	result, err := database.DB.Exec(
		"UPDATE students SET name = ?, tel = ?, study = ? WHERE id = ?",
		updateData.Name, updateData.Tel, updateData.Study, id,
	)

	if err != nil {
		log.Printf("更新学生 %s 失败: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新学生失败"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("获取影响行数失败: %v", err)
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
		return
	}

	log.Printf("成功更新学生 %s, 影响行数: %d", id, rowsAffected)
	c.JSON(http.StatusOK, gin.H{"status": "更新成功"})
}

// 删除学生
func DeleteStudent(c *gin.Context) {
	id := c.Param("id")

	result, err := database.DB.Exec(
		"DELETE FROM students WHERE id = ?", id,
	)

	if err != nil {
		log.Printf("删除学生 %s 失败: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除学生失败"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("获取影响行数失败: %v", err)
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
		return
	}

	log.Printf("成功删除学生 %s, 影响行数: %d", id, rowsAffected)
	c.JSON(http.StatusOK, gin.H{"status": "删除成功"})
}
