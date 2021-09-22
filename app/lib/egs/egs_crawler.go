package egs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const egs_url = "https://store-site-backend-static-ipv4.ak.epicgames.com/freeGamesPromotions?locale=ru&country=UA&allowCountries=UA"

type EGSGame struct {
	ID          string
	URL         string
	Description string
	Title       string
}

func ParseFreeEgsGamesUrls() (map[string]*EGSGame, error) {
	res, err := http.Get(egs_url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		message := fmt.Sprintf("status code error: %d %s", res.StatusCode, res.Status)
		err = fmt.Errorf(message)
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)

	var result AutoGeneratedEGSResponse
	if err = json.Unmarshal(body, &result); err != nil {
		fmt.Println("Can not unmarshal JSON")
		return nil, err
	}
	games := map[string]*EGSGame{}
	elements := result.Data.Catalog.SearchStore.Elements
	for _, element := range elements {
		offers := element.Promotions.PromotionalOffers
		if offers == nil {
			continue
		}

		for _, offerParent := range offers {
			for _, offer := range offerParent.PromotionalOffers {
				now := time.Now()
				if offer.StartDate.Before(now) && offer.EndDate.After(now) {
					if offer.DiscountSetting.DiscountPercentage == 0 {
						productUrl := "https://www.epicgames.com/store/ru/p/" + element.ProductSlug
						games[element.ID] = &EGSGame{element.ID, productUrl, element.Description, element.Title}
					}

				}
			}
		}
	}

	return games, nil
}
