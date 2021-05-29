package productRepository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/korableg/getproduct/pkg/product"
	"github.com/korableg/getproduct/pkg/productLocalProvider"
	"github.com/korableg/getproduct/pkg/productProvider"
)

type ProductRepository struct {
	providers     []productProvider.ProductProvider
	localProvider productLocalProvider.ProductLocalProvider
	muProviders   sync.RWMutex
}

func New(localProvider productLocalProvider.ProductLocalProvider) *ProductRepository {
	pr := ProductRepository{
		providers:     make([]productProvider.ProductProvider, 0, 10),
		localProvider: localProvider,
		muProviders:   sync.RWMutex{},
	}

	return &pr
}

func (pr *ProductRepository) AddProvider(provider productProvider.ProductProvider) {
	pr.muProviders.Lock()
	defer pr.muProviders.Unlock()
	pr.providers = append(pr.providers, provider)
}

func (pr *ProductRepository) Get(ctx context.Context, barcode string) (*product.Product, error) {
	pr.muProviders.RLock()
	defer pr.muProviders.RUnlock()

	err := pr.checkProviders()
	if err != nil {
		return nil, err
	}

	log.Println(fmt.Sprintf("getting product by barcode: %s", barcode))

	newCtx, cancelFunc := context.WithTimeout(ctx, time.Second*10)
	defer cancelFunc()

	productChan := make(chan *product.Product)
	fetchingDoneChan := make(chan struct{})

	pr.getProductWithProviders(newCtx, barcode, productChan, fetchingDoneChan)

	select {
	case dst, ok := <-productChan:
		if ok {
			return dst, nil
		} else {
			return nil, fmt.Errorf("product by barcode %s not found", barcode)
		}
	case <-fetchingDoneChan:
		return nil, fmt.Errorf("product by barcode %s not found", barcode)
	case <-newCtx.Done():
		return nil, newCtx.Err()
	}

}

func (pr *ProductRepository) GetTheBest(ctx context.Context, barcode string) (*product.Product, error) {

	if pr.localProvider != nil {
		p, _ := pr.localProvider.GetProduct(ctx, barcode)
		if p != nil {
			return p, nil
		}
	}

	pr.muProviders.RLock()
	defer pr.muProviders.RUnlock()

	err := pr.checkProviders()
	if err != nil {
		return nil, err
	}

	log.Println(fmt.Sprintf("getting the best matches product by barcode: %s", barcode))

	newCtx, cancelFunc := context.WithTimeout(ctx, time.Second*10)
	defer cancelFunc()

	products := make([]*product.Product, 0, len(pr.providers))
	productChan := make(chan *product.Product)

	fetchingDoneChan := make(chan struct{})

	pr.getProductWithProviders(newCtx, barcode, productChan, fetchingDoneChan)

	for {
		select {
		case dst, ok := <-productChan:
			if ok {
				products = append(products, dst)
			} else {
				if len(products) > 0 {
					p := pr.chooseTheBestProduct(products)
					if pr.localProvider != nil {
						pr.localProvider.AddProduct(ctx, p)
					}
					return p, nil
				} else {
					return nil, fmt.Errorf("product by barcode %s not found", barcode)
				}
			}

		case <-newCtx.Done():
			if len(products) > 0 {
				return pr.chooseTheBestProduct(products), nil
			}
			return nil, newCtx.Err()
		}
	}

}

func (pr *ProductRepository) GetAll(ctx context.Context, barcode string) ([]*product.Product, error) {

	pr.muProviders.RLock()
	defer pr.muProviders.RUnlock()

	err := pr.checkProviders()
	if err != nil {
		return nil, err
	}

	log.Println(fmt.Sprintf("getting all products by barcode: %s", barcode))

	newCtx, cancelFunc := context.WithTimeout(ctx, time.Second*10)
	defer cancelFunc()

	products := make([]*product.Product, 0, len(pr.providers))
	productChan := make(chan *product.Product)

	fetchingDoneChan := make(chan struct{})

	pr.getProductWithProviders(newCtx, barcode, productChan, fetchingDoneChan)

	for {
		select {
		case dst, ok := <-productChan:
			if ok {
				products = append(products, dst)
			} else {
				if len(products) > 0 {
					p := pr.chooseTheBestProduct(products)
					if pr.localProvider != nil {
						pr.localProvider.AddProduct(ctx, p)
					}
					return products, nil
				} else {
					return nil, fmt.Errorf("product by barcode %s not found", barcode)
				}
			}

		case <-newCtx.Done():
			if len(products) > 0 {
				return products, nil
			}
			return nil, newCtx.Err()
		}

	}

}

func (pr *ProductRepository) DeleteFromLocalProvider(ctx context.Context, barcode string) error {

	if pr.localProvider == nil {
		return nil
	}

	return pr.localProvider.DeleteProduct(ctx, barcode)

}

func (pr *ProductRepository) getProductWithProviders(
	ctx context.Context, barcode string, productChan chan<- *product.Product, fetchingDoneChan chan<- struct{}) {

	wg := &sync.WaitGroup{}
	wg.Add(len(pr.providers))

	go func() {
		wg.Wait()
		fetchingDoneChan <- struct{}{}
		close(productChan)
	}()

	for _, provider := range pr.providers {
		go func(provider productProvider.ProductProvider, wg *sync.WaitGroup) {
			p, err := provider.GetProduct(ctx, barcode)
			defer wg.Done()
			if err != nil {
				log.Println(err)
				return
			}
			if p != nil {
				productChan <- p
			}

		}(provider, wg)
	}

}

func (pr *ProductRepository) checkProviders() error {
	if len(pr.providers) == 0 {
		return errors.New("product providers is empty")
	}

	return nil
}

func (pr *ProductRepository) chooseTheBestProduct(products []*product.Product) *product.Product {

	winner := products[0]
	winnerScore := winner.Rating()

	for i := 1; i < len(products); i++ {
		winnerCandidateRating := products[i].Rating()
		if winnerCandidateRating > winnerScore {
			winner = products[i]
			winnerScore = winnerCandidateRating
		}
	}

	return winner

}
