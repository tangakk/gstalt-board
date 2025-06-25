switch (window.location.pathname) {
	case "/":
		const result = document.getElementById("result")
		if (result.childElementCount > 0) {
			const template = result.children[0].cloneNode(true);
			result.innerHTML = "";
			fetch("/api/get/boards").then(res => {
				res.json().then(jsonBody => { 
					for (jsonElem of jsonBody) {
						const tmp = template.cloneNode(true);
						tmp.getElementsByClassName("board-name")[0].innerHTML = jsonElem.Name;
						tmp.getElementsByClassName("board-desc")[0].innerHTML = jsonElem.Description;
						tmp.setAttribute("href", jsonElem.Name)
						result.append(tmp)
					}
				})
			})
		}
		break;
}