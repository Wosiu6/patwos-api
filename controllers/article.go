package controllers

import (
	"net/http"
	"strconv"

	"github.com/Wosiu6/patwos-api/models"
	"github.com/Wosiu6/patwos-api/service"
	"github.com/gin-gonic/gin"
)

type ArticleController struct {
	service service.ArticleService
}

func NewArticleController(articleService service.ArticleService) *ArticleController {
	return &ArticleController{service: articleService}
}

func (ac *ArticleController) GetArticles(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	articles, err := ac.service.GetAllArticles(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch articles"})
		return
	}

	summaries := make([]models.ArticleSummaryResponse, 0, len(articles))
	for _, article := range articles {
		summaries = append(summaries, models.ArticleSummaryResponse{
			ID:        article.ID,
			Title:     article.Title,
			Slug:      article.Slug,
			Author:    article.Author,
			CreatedAt: article.CreatedAt,
			UpdatedAt: article.UpdatedAt,
			Views:     article.Views,
		})
	}

	c.JSON(http.StatusOK, gin.H{"articles": summaries})
}

func (ac *ArticleController) GetArticle(c *gin.Context) {
	id := c.Param("id")

	articleID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		slug := id
		article, err := ac.service.GetArticleBySlug(slug)
		if err != nil {
			if err == service.ErrArticleNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch article"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"article": article.ToResponse()})
		return
	}

	article, err := ac.service.GetArticle(uint(articleID))
	if err != nil {
		if err == service.ErrArticleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch article"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"article": article.ToResponse()})
}

func (ac *ArticleController) CreateArticle(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article, err := ac.service.CreateArticle(req.Title, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create article"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"article": article.ToResponse()})
}

func (ac *ArticleController) UpdateArticle(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article, err := ac.service.UpdateArticle(uint(articleID), req.Title, userID.(uint))
	if err != nil {
		if err == service.ErrArticleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		if err == service.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only edit your own articles"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update article"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"article": article.ToResponse()})
}

func (ac *ArticleController) DeleteArticle(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err = ac.service.DeleteArticle(uint(articleID), userID.(uint))
	if err != nil {
		if err == service.ErrArticleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		if err == service.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own articles"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete article"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Article deleted successfully"})
}

func (ac *ArticleController) GetArticleViews(c *gin.Context) {
	id := c.Param("id")

	// Try numeric ID first; fallback to slug
	articleID, err := strconv.ParseUint(id, 10, 32)
	var article *serviceArticle
	if err != nil {
		// Treat as slug
		a, err := ac.service.GetArticleBySlug(id)
		if err != nil {
			if err == service.ErrArticleNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch article"})
			return
		}
		article = &serviceArticle{ID: a.ID}
	} else {
		a, err := ac.service.GetArticle(uint(articleID))
		if err != nil {
			if err == service.ErrArticleNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch article"})
			return
		}
		article = &serviceArticle{ID: a.ID}
	}

	views, err := ac.service.GetArticleViews(article.ID)
	if err != nil {
		if err == service.ErrArticleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch views"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"views": views})
}

func (ac *ArticleController) IncrementArticleViews(c *gin.Context) {
	id := c.Param("id")

	// Try numeric ID first; fallback to slug
	articleID, err := strconv.ParseUint(id, 10, 32)
	var article *serviceArticle
	if err != nil {
		// Treat as slug
		a, err := ac.service.GetArticleBySlug(id)
		if err != nil {
			if err == service.ErrArticleNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch article"})
			return
		}
		article = &serviceArticle{ID: a.ID}
	} else {
		a, err := ac.service.GetArticle(uint(articleID))
		if err != nil {
			if err == service.ErrArticleNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch article"})
			return
		}
		article = &serviceArticle{ID: a.ID}
	}

	views, err := ac.service.IncrementArticleViews(article.ID)
	if err != nil {
		if err == service.ErrArticleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to increment views"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"views": views})
}

type serviceArticle struct {
	ID uint
}
