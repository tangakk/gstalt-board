const path = window.location.pathname;
let result;
switch (true) {
	case (path.match("^/$") !== null):
		result = document.getElementById("result")
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
				}).catch(err => {
					console.error("Ошибка при шаблонизации: "+err);
				})
			}).catch(err => {
				console.error("Ошибка при запросе: "+err);
			})
		} else {
			console.error("Нет шаблона!");
		}
		break;
	case (path.match("^/.*$") !== null):
		document.getElementById("board-title").innerHTML = path;
		result = document.getElementById("result")
		if (result.childElementCount > 0) {
			const template = result.children[0].cloneNode(true);
			result.innerHTML = "";
			fetch("/api/get/posts?board="+path.slice(1)+"&from=0&to=20").then(res => {
				res.json().then(jsonBody => { 
					for (jsonElem of jsonBody) {
						const tmp = template.cloneNode(true);
						tmp.getElementsByClassName("post-id")[0].innerHTML = "#"+jsonElem.Id;
						tmp.getElementsByClassName("post-author")[0].innerHTML = jsonElem.Author;
						tmp.getElementsByClassName("post-text")[0].innerHTML = jsonElem.Text;
						tmp.getElementsByClassName("post-time")[0].innerHTML = jsonElem.Timestamp; // FIXME: Конвертировать в человекочитаемый формат
						tmp.getElementsByClassName("post-reply")[0].setAttribute("href", path+"/"+jsonElem.Id);
						result.append(tmp)
					}
				}).catch(err => {
					console.error("Ошибка при шаблонизации: "+err);
				})
			}).catch(err => {
				console.error("Ошибка при запросе: "+err);
			})
		} else {
			console.error("Нет шаблона!");
		}
		break;
}