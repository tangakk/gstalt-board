async function rootTemplating () {
	const path = window.location.pathname;
	const result = document.getElementById("result")
	if (result.childElementCount > 0) {
		try {
			const template = result.children[0].cloneNode(true);
			result.innerHTML = "";
			const jsonBody = await (await fetch("/api/get/boards")).json();
			for (jsonElem of jsonBody) {
				const tmp = template.cloneNode(true);
				tmp.getElementsByClassName("board-name")[0].innerHTML = jsonElem.Name;
				tmp.getElementsByClassName("board-desc")[0].innerHTML = jsonElem.Description;
				tmp.setAttribute("href", jsonElem.Name)
				result.append(tmp)
			}
		} catch (err) {
			console.error("Ошибка выполнения: "+err);
		}
	} else {
		console.error("Нет шаблона!");
	}
}
async function boardTemplating () {
	const path = window.location.pathname;
	document.getElementById("board-title").innerHTML = path;
	const result = document.getElementById("result");
	document.getElementById("form-board").setAttribute("value", path);
	if (result.childElementCount > 0) {
		try {
			const template = result.children[0].cloneNode(true);
			result.innerHTML = "";
			const postCount = await (await fetch(`/api/get/posts?board=${path.slice(1)}`)).json();
			const jsonBody = await (await fetch(`/api/get/posts?board=${path.slice(1)}&from=${postCount.length-3}&to=${postCount.length}`)).json();
			for (const jsonElem of jsonBody) {
				const tmp = template.cloneNode(true);
				tmp.getElementsByClassName("post-id")[0].innerHTML = "#"+jsonElem.Id;
				tmp.getElementsByClassName("post-author")[0].innerHTML = jsonElem.Author;
				tmp.getElementsByClassName("post-text")[0].innerHTML = jsonElem.Text;
				tmp.getElementsByClassName("post-time")[0].innerHTML = jsonElem.Timestamp; // FIXME: Конвертировать в человекочитаемый формат
				tmp.getElementsByClassName("post-reply")[0].setAttribute("href", path+"/"+jsonElem.Id);
				result.prepend(tmp);
			}
		} catch (err) {
			console.error("Ошибка выполнения: "+err);
		}
	} else {
		console.error("Нет шаблона!");
	}
}
async function postTemplating () {}

const path = window.location.pathname;
switch (true) {
	case (path.match("^/$") !== null):
		window.onload = rootTemplating;
		break;
	case (path.match("^/.*$") !== null):
		window.onload = boardTemplating;
		break;
}