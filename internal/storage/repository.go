package storage

import "errors"

var databaseURL []RedirectURL = []RedirectURL{}

func RepositoryAddURL(inputURL string) RedirectURL {

	var newURL RedirectURL = RedirectURL{
		Id:      len(databaseURL),
		Address: inputURL,
	}

	databaseURL = append(databaseURL, newURL)
	return newURL
}

func RepositoryGetURLById(id int) (RedirectURL, error) {
	var newURL RedirectURL

	for _, v := range databaseURL {
		if v.Id == id {
			newURL = v
		}
	}

	if newURL.Address == "" {
		return newURL, errors.New("DB doesn`t have value")
	}

	return newURL, nil

}

//type Product struct {
//	Id          string
//	Name        string
//	Description string
//	Price       float64
//	Stock       int
//}

//type ProductInput struct {
//	Name        string
//	Description string
//	Price       float64
//	Stock       int
//}

//func AddProduct(productInput model.ProductInput) model.Product {
//	var newProduct model.Product = model.Product{
//		Id:          uuid.NewString(),
//		Name:        productInput.Name,
//		Description: productInput.Description,
//		Price:       productInput.Price,
//		Stock:       productInput.Stock,
//	}
//
//	database = append(database, newProduct)
//	return newProduct
//}

//func GetProductById(id string) (int, model.Product) {
//	var product model.Product
//	var productIndex int
//
//	for index, v := range database {
//		if v.Id == id {
//			product = v
//			productIndex = index
//		}
//	}
//
//	return productIndex, product
//}
