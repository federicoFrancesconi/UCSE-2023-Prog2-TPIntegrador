package services

import (
	"UCSE-2023-Prog2-TPIntegrador/dto"
	"UCSE-2023-Prog2-TPIntegrador/model"
	"UCSE-2023-Prog2-TPIntegrador/repositories"
	"UCSE-2023-Prog2-TPIntegrador/utils"
	"errors"
)

type EnvioServiceInterface interface {
	CrearEnvio(*dto.Envio) error
	ObtenerEnviosFiltrados(utils.FiltroEnvio) ([]*dto.Envio, error)
	ObtenerEnvioPorId(*dto.Envio) (*dto.Envio, error)
	ObtenerBeneficioEntreFechas(utils.FiltroEnvio) (float32, error)
	AgregarParada(*dto.Envio) (bool, error)
	IniciarViaje(*dto.Envio) (bool, error)
	FinalizarViaje(*dto.Envio) (bool, error)
}

type EnvioService struct {
	envioRepository    repositories.EnvioRepositoryInterface
	camionRepository   repositories.CamionRepositoryInterface
	pedidoRepository   repositories.PedidoRepositoryInterface
	productoRepository repositories.ProductoRepositoryInterface
}

func NewEnvioService(envioRepository repositories.EnvioRepositoryInterface, camionRepository repositories.CamionRepositoryInterface, pedidoRepository repositories.PedidoRepositoryInterface, productoRepository repositories.ProductoRepositoryInterface) *EnvioService {
	return &EnvioService{
		envioRepository:    envioRepository,
		camionRepository:   camionRepository,
		pedidoRepository:   pedidoRepository,
		productoRepository: productoRepository,
	}
}

func (service *EnvioService) CrearEnvio(envio *dto.Envio) error {
	envioCabeEnCamion, err := service.envioCabeEnCamion(envio)

	if err != nil {
		return err
	}

	if !envioCabeEnCamion {
		//Devuelve un error diciendo que el envio no cabe en el camion
		return errors.New("el envio no cabe en el camion")
	}

	//al crearlo coloco el envio en estado despachar
	envio.Estado = model.EstadoEnvio(model.ParaEnviar)

	//Cambio el estado de los pedidos del envio
	err = service.enviarPedidosDeEnvio(envio)

	if err != nil {
		return err
	}

	return service.envioRepository.CrearEnvio(envio.GetModel())
}

func (service *EnvioService) ObtenerEnviosFiltrados(filtroEnvio utils.FiltroEnvio) ([]*dto.Envio, error) {
	enviosDB, err := service.envioRepository.ObtenerEnviosFiltrados(filtroEnvio)

	if err != nil {
		return nil, err
	}

	var envios []*dto.Envio
	for _, envioDB := range enviosDB {
		envio := dto.NewEnvio(envioDB)
		envios = append(envios, envio)
	}
	return envios, nil
}

func (service *EnvioService) ObtenerEnvioPorId(envioConID *dto.Envio) (*dto.Envio, error) {
	envioDB, err := service.envioRepository.ObtenerEnvioPorId(envioConID.GetModel())

	var envio *dto.Envio

	if err != nil {
		return nil, err
	} else {
		envio = dto.NewEnvio(envioDB)
	}

	return envio, nil
}

func (service *EnvioService) envioCabeEnCamion(envio *dto.Envio) (bool, error) {
	//Primero buscamos el camion por patente
	camionConPatente := model.Camion{Patente: envio.PatenteCamion}
	camion, err := service.camionRepository.ObtenerCamionPorPatente(camionConPatente)

	if err != nil {
		return false, err
	}

	//Obtenemos el peso total de los pedidos
	var pesoTotal float32 = 0
	for _, idPedido := range envio.Pedidos {
		pedido, err := service.pedidoRepository.ObtenerPedidoPorId(model.Pedido{Id: idPedido})

		if err != nil {
			return false, err
		}

		//Calculo el peso del pedido sumando el peso de cada producto elegido
		var peso float32 = 0
		for _, producto := range pedido.ProductosElegidos {
			peso += producto.ObtenerPesoProductoPedido()
		}

		pesoTotal += peso
	}

	//Verificamos si el peso total de los pedidos es menor o igual al peso maximo del camion
	if pesoTotal <= float32(camion.PesoMaximo) {
		return true, nil
	} else {
		return false, nil
	}
}

func (service *EnvioService) enviarPedidosDeEnvio(envio *dto.Envio) error {
	for _, idPedido := range envio.Pedidos {
		err := service.enviarPedido(&dto.Pedido{Id: idPedido})
		if err != nil {
			return err
		}
	}
	return nil
}

func (service *EnvioService) enviarPedido(pedidoPorEnviar *dto.Pedido) error {
	//Primero buscamos el pedido a enviar
	pedido, err := service.pedidoRepository.ObtenerPedidoPorId(pedidoPorEnviar.GetModel())

	if err != nil {
		return err
	}

	//Valida que el pedido esté en estado Aceptado
	if pedido.Estado != model.Aceptado {
		return nil
	}

	//Cambia el estado del pedido a Para enviar, si es que no estaba ya en ese estado
	if pedido.Estado != model.ParaEnviar {
		pedido.Estado = model.ParaEnviar
	}

	//Actualiza el pedido en la base de datos
	return service.pedidoRepository.ActualizarPedido(*pedido)
}

func (service *EnvioService) ObtenerBeneficioEntreFechas(filtro utils.FiltroEnvio) (float32, error) {
	//Busca los envios entre el rango de fechas pasado como parametro
	filtroEnvio := utils.FiltroEnvio{
		PatenteCamion:         "",
		Estado:                -1,
		UltimaParada:          "",
		FechaCreacionComienzo: filtro.FechaCreacionComienzo,
		FechaCreacionFin:      filtro.FechaCreacionFin,
	}

	envios, err := service.ObtenerEnviosFiltrados(filtroEnvio)

	if err != nil {
		return 0, err
	}

	//Suma el precio de los pedidos de cada envio
	var beneficioBruto float32 = 0
	for _, envio := range envios {
		precioTotal, err := service.getPrecioTotalProductosDeEnvio(envio)
		
		if err != nil {
			return 0, err
		}

		beneficioBruto += precioTotal
	}

	//Suma el costo de los envios
	var costoEnvios float32 = 0
	for _, envio := range envios {
		costoEnvio, err := service.getCostoEnvio(envio)

		if err != nil {
			return 0, err
		}

		costoEnvios += costoEnvio
	}

	beneficioNeto := beneficioBruto - costoEnvios

	return beneficioNeto, nil
}

func (service *EnvioService) getPrecioTotalProductosDeEnvio(envio *dto.Envio) (float32, error) {
	var precioTotal float32 = 0
	
	for _, idPedido := range envio.Pedidos {
		pedido, err := service.pedidoRepository.ObtenerPedidoPorId(model.Pedido{Id: idPedido})

		if err != nil {
			return 0, err
		}

		//Convierto el pedido a dto
		pedidoDTO := dto.NewPedido(pedido)

		precioTotal += pedidoDTO.ObtenerPecioTotal()
	}

	return precioTotal, nil
}

func (service *EnvioService) getCostoEnvio(envio *dto.Envio) (float32, error) {
	//Obtiene el camion del envio para conocer el costoPorKilometro
	camion, err := service.camionRepository.ObtenerCamionPorPatente(model.Camion{Patente: envio.PatenteCamion})

	if err != nil {
		return 0, err
	}

	//Suma los kilometros de cada parada
	var kilometrosRecorridos int = 0
	for i := 0; i < len(envio.Paradas)-1; i++ {
		//Obtiene la distancia entre la parada i y la parada i+1
		kilometrosRecorridos += envio.Paradas[i].KmRecorridos
	}

	costoEnvio := camion.CostoPorKilometro * float32(kilometrosRecorridos)

	return costoEnvio, nil
}

func (service *EnvioService) AgregarParada(envio *dto.Envio) (bool, error) {
	if envio.Estado != model.EnRuta {
		return false, errors.New("el envio no esta en ruta")
	}

	return true, service.envioRepository.ActualizarEnvio(envio.GetModel())
}

func (service *EnvioService) IniciarViaje(envio *dto.Envio) (bool, error) {
	if envio.Estado != model.ADespachar {
		return false, nil
	}

	envio.Estado = model.EnRuta

	return true, service.envioRepository.ActualizarEnvio(envio.GetModel())
}

func (service *EnvioService) FinalizarViaje(envio *dto.Envio) (bool, error) {
	if envio.Estado == model.Despachado {
		return false, nil
	}

	envio.Estado = model.Despachado

	service.envioRepository.ActualizarEnvio(envio.GetModel())

	//pasar pedidos a estado enviado
	service.entregarPedidosDeEnvio(envio)

	//descontar stock de productos
	service.descontarStockProductosDeEnvio(envio)

	return true, nil
}

func (service *EnvioService) entregarPedidosDeEnvio(envio *dto.Envio) error {
	for _, idPedido := range envio.Pedidos {

		//Descuenta el stock de los productos
		err := service.entregarPedido(&dto.Pedido{Id: idPedido})

		if err != nil {
			return err
		}
	}
	return nil
}

func (service *EnvioService) entregarPedido(pedidoPorEntregar *dto.Pedido) error {
	//Primero buscamos el pedido a entregar
	pedido, err := service.pedidoRepository.ObtenerPedidoPorId(pedidoPorEntregar.GetModel())

	if err != nil {
		return err
	}

	//Valida que el pedido esté en estado Para enviar
	if pedido.Estado != model.ParaEnviar {
		return nil
	}

	//Cambia el estado del pedido a Enviado, si es que no estaba ya en ese estado
	if pedido.Estado != model.Enviado {
		pedido.Estado = model.Enviado
	}

	//Actualiza el pedido en la base de datos
	return service.pedidoRepository.ActualizarPedido(*pedido)
}

func (service *EnvioService) descontarStockProductosDeEnvio(envio *dto.Envio) error {
	for _, idPedido := range envio.Pedidos {
		pedido, err := service.pedidoRepository.ObtenerPedidoPorId(model.Pedido{Id: idPedido})
		if err != nil {
			return err
		}

		for _, producto := range pedido.ProductosElegidos {
			err = service.descontarStockProducto(*dto.NewProductoPedido(&producto))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (service *EnvioService) descontarStockProducto(productoPedido dto.ProductoPedido) error {
	//Generamos un producto con el codigo del producto del pedido
	productoConId := model.Producto{CodigoProducto: productoPedido.CodigoProducto}

	//Buscamos el producto del que hay que descontar la cantidad
	producto, err := service.productoRepository.ObtenerProductoPorCodigo(productoConId)

	if err != nil {
		return err
	}

	//Modificamos el stock
	producto.StockActual = producto.StockActual - productoPedido.Cantidad

	//Actualizamos la base de datos
	return service.productoRepository.ActualizarProducto(*producto)
}