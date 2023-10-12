package handlers

import (
	"UCSE-2023-Prog2-TPIntegrador/dto"
	"UCSE-2023-Prog2-TPIntegrador/services"
	"UCSE-2023-Prog2-TPIntegrador/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProductoHandler struct {
	productoService services.ProductoServiceInterface
}

func NewProductoHandler(productoService services.ProductoServiceInterface) *ProductoHandler {
	return &ProductoHandler{productoService: productoService}
}

func (handler *ProductoHandler) ObtenerProductos(c *gin.Context) {
	user := dto.NewUser(utils.GetUserInfoFromContext(c))

	productos, err := handler.productoService.ObtenerProductos()

	//Si hay un error, lo devolvemos
	if err != nil {
		log.Printf("[handler:ProductoHandler][method:ObtenerProductos][error:%s][user:%s]", err.Error(), user.Codigo)

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Agregamos un log para indicar información relevante del resultado
	log.Printf("[handler:ProductoHandler][method:ObtenerProductos][cantidad:%d][user:%s]", len(productos), user.Codigo)

	c.JSON(http.StatusOK, productos)
}

func (handler *ProductoHandler) CrearProducto(c *gin.Context) {
	user := dto.NewUser(utils.GetUserInfoFromContext(c))

	var producto dto.Producto

	//Parseamos el body del request y lo guardamos en el objeto producto
	if err := c.ShouldBindJSON(&producto); err != nil {
		log.Printf("[handler:ProductoHandler][method:CrearProducto][error:%s][user:%s]", err.Error(), user.Codigo)

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Creamos el producto en la base de datos
	if err := handler.productoService.CrearProducto(&producto); err != nil {
		log.Printf("[handler:ProductoHandler][method:CrearProducto][error:%s][user:%s]", err.Error(), user.Codigo)

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Agregamos un log para indicar información relevante del resultado
	log.Printf("[handler:ProductoHandler][method:CrearProducto][user:%s]", user.Codigo)

	c.JSON(http.StatusCreated, producto)
}

// Handler para eliminar un producto
func (handler *ProductoHandler) EliminarProducto(c *gin.Context) {
	user := dto.NewUser(utils.GetUserInfoFromContext(c))

	var producto dto.Producto

	//Parseamos el body del request y lo guardamos en el objeto producto
	if err := c.ShouldBindJSON(&producto); err != nil {
		log.Printf("[handler:ProductoHandler][method:EliminarProducto][error:%s][user:%s]", err.Error(), user.Codigo)

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Eliminamos el producto de la base de datos
	if err := handler.productoService.EliminarProducto(&producto); err != nil {
		log.Printf("[handler:ProductoHandler][method:EliminarProducto][error:%s][user:%s]", err.Error(), user.Codigo)

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Agregamos un log para indicar información relevante del resultado
	log.Printf("[handler:ProductoHandler][method:EliminarProducto][user:%s]", user.Codigo)

	c.JSON(http.StatusOK, producto)
}