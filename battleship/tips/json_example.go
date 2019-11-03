
	// JSON struct
	type JsonResponse struct {
		ChallengerID 	int 		`json:"challengerID"`
		NextPage		string		`json:"redirect"`
		Time			time.Time	`json:"timestamp"`
	}

	// JSON string
	var j string
		p, err := app.players.Get(challenger.ID)
		if err != nil {
			if xerrors.Is(err, models.ErrNoRecord) {
				app.notFound(w)
			} else {
				app.serverError(w, err)
			}
			return
		}
		j += fmt.Sprintf(`{"challengerID":%d, "redirect":"/status/battles/%d"}`, challenger.ID)
		bytes := []byte(j)

		var JR JsonResponse
		JR.Time = time.Now()
		err = json.Unmarshal(bytes, &JR)
		if err != nil {
			app.serverError(w, err)
			return
		}
		out, err := json.Marshal(JR)
		if err != nil {
			app.serverError(w, err)
			return
		}
		app.renderJson(w, r, out)
	}
