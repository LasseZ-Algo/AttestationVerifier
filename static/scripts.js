document.addEventListener('DOMContentLoaded', (event) => {
    const dropArea = document.querySelector('.file-drop-area');
    const fileInput = document.querySelector('#fileInput');
    const fakeBtn = document.querySelector('.fake-btn');
    const fileMsg = document.querySelector('.file-msg');

    // Event listeners for drag and drop
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropArea.addEventListener(eventName, preventDefaults, false);
    });

    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    ['dragenter', 'dragover'].forEach(eventName => {
        dropArea.addEventListener(eventName, () => dropArea.classList.add('is-active'), false);
    });

    ['dragleave', 'drop'].forEach(eventName => {
        dropArea.addEventListener(eventName, () => dropArea.classList.remove('is-active'), false);
    });

    dropArea.addEventListener('drop', handleDrop, false);

    function handleDrop(e) {
        let dt = e.dataTransfer;
        let files = dt.files;
        handleFiles(files);
    }

    function handleFiles(files) {
        ([...files]).forEach(uploadFile);
    }

    function updateFileDisplay() {
        if (fileInput.files.length > 0) {
            const fileName = fileInput.files[0].name;
            fakeBtn.textContent = fileName;
            fileMsg.textContent = "Click again to change " + fileName;
            fakeBtn.classList.add('remove-btn');
        }
    }
    
    fileInput.addEventListener('change', (e) => {
        updateFileDisplay();
    });
    
    function uploadFile(file) {
        const dataTransfer = new DataTransfer();
        dataTransfer.items.add(file);
        fileInput.files = dataTransfer.files;
        updateFileDisplay();
    }

    fakeBtn.addEventListener('click', (e) => {
        e.preventDefault(); // Prevent default behavior which might cause double opening
        fileInput.click();
    });


    fileInput.addEventListener('change', (e) => {
        if (fileInput.files.length > 0) {
            const fileName = fileInput.files[0].name;
            fakeBtn.textContent = fileName;
            fileMsg.textContent = "Click again to change " + fileName;
            fakeBtn.classList.add('remove-btn');
        }
    });

    // Form submission via fetch
    document.getElementById('verifyForm').addEventListener('submit', function(e) {
        e.preventDefault();
        fetch('/verify', {
            method: 'POST',
            body: new FormData(this)
        })
        .then(response => response.json())
        .then(data => {
            const resultDiv = document.getElementById('result');
            resultDiv.textContent = data.message;
            resultDiv.className = `result ${data.message.includes('failed') ? 'error' : 'success'}`;
            resultDiv.style.display = 'block';

            const reportDetails = document.getElementById('reportDetails');
            if (data.report_details) {
                document.getElementById('source').textContent = data.report_details.Source;
                document.getElementById('protocol').textContent = data.report_details.Protocol;
                document.getElementById('product').textContent = data.report_details.Product;
                reportDetails.style.display = 'block';
            } else {
                reportDetails.style.display = 'none';
            }
        })
        .catch(error => {
            console.error('Error:', error);
            document.getElementById('result').textContent = 'An error occurred.';
            document.getElementById('result').className = 'result error';
            document.getElementById('result').style.display = 'block';
        });
    });
});