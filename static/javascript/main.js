console.log('Start JS')

const API_ENDPOINTS = {
	WordCount: "/api/wordcount"
};

function toggleLoading() {
	document.querySelector('.container').classList.toggle('loading')
}

function uploadFile(fileInputElement) {
	toggleLoading()
	var formData = new FormData();

	formData.append("usertxtfile", fileInputElement.files[0]);

	var request = new XMLHttpRequest();
	request.open("POST", API_ENDPOINTS.WordCount);
	request.send(formData);

	request.addEventListener('load', (ev) => {
		if (request.status == 200) {
			uploaded(request.responseText)
		} else {
			console.error(ev)
		}

	});

	request.addEventListener('error', (ev) => {
		alert('[ERROR] Errors occurred. Use Chrome debugger to see.')
		console.error(ev)

	});

	request.addEventListener('loadend', (ev) => {
		toggleLoading()
	});
}

function uploaded(res) {
	document.querySelector('.container .result').innerHTML = `The word count is: ${res}`
}