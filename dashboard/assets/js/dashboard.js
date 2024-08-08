let page = 1;
let totalPages = 0;

function fetchData(collectionName, count, currentPage = 1) {
    page = currentPage;

    fetch(`/api/v1/docs/${collectionName}?page=${page}`)
        .then(response => response.json())
        .then(data => {
            const recordsTable = document.getElementById('table-body');
            const respTime = document.getElementById('collection_fetch_time');
            const recCount = document.getElementById('record_count');
            const collectionsTitle = document.getElementById('collection_title');
            const pageCount = document.getElementById('page_count');

            totalPages = Math.ceil(count / 5);

            pageCount.innerText = Math.ceil(count / 5);
            collectionsTitle.innerText = collectionName;
            respTime.innerText = data._resp;
            recCount.innerText = count;
            recordsTable.innerHTML = '';

            data.documents.forEach(doc => {
                const tr = document.createElement('tr');
                tr.id = `record-${doc.id}`;

                const fileTypeTd = document.createElement('td');
                fileTypeTd.className = 'px-1 py-2 border-b sm:p-3 border-main';
                fileTypeTd.innerHTML = `
                    <div class='flex items-center'>
                        <svg class='p-1.5 mr-2.5 w-7 h-7 rounded-lg border border-main bg-sec' xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round' class='lucide lucide-file-json-2'>
                            <path d='M4 22h14a2 2 0 0 0 2-2V7l-5-5H6a2 2 0 0 0-2 2v4' />
                            <path d='M14 2v4a2 2 0 0 0 2 2h4' />
                            <path d='M4 12a1 1 0 0 0-1 1v1a1 1 0 0 1-1 1 1 1 0 0 1 1 1v1a1 1 0 0 0 1 1' />
                            <path d='M8 18a1 1 0 0 0 1-1v-1a1 1 0 0 1 1-1 1 1 0 0 1-1-1v-1a1 1 0 0 0-1-1' />
                        </svg>
                        JSON
                    </div>
                `;

                const idTd = document.createElement('td');
                idTd.className = 'px-1 py-2 border-b sm:p-3 border-main';
                idTd.textContent = doc.id;

                const contentTd = document.createElement('td');
                contentTd.className = 'hidden px-1 py-2 border-b sm:p-3 border-main md:table-cell';

                if (doc.length > 5) {
                    contentTd.innerHTML = `<code>${JSON.stringify(doc)}</code>`;
                } else {
                    contentTd.innerHTML = `<code>${Object.keys(doc.data).length}</code>`;
                }

                const dateAddedTd = document.createElement('td');
                dateAddedTd.className = 'px-1 py-2 border-b sm:p-3 border-main';
                dateAddedTd.textContent = new Date().toLocaleString();

                tr.appendChild(fileTypeTd);
                tr.appendChild(idTd);
                tr.appendChild(contentTd);
                tr.appendChild(dateAddedTd);

                idTd.addEventListener('click', () => {
                    fetchDocumentDetails(collectionName, doc.id);
                });

                recordsTable.appendChild(tr);
            });
        })
        .catch(error => {
            console.error('Error fetching data:', error);
        });
}

function fetchDocumentDetails(collectionName, documentId) {
    fetch(`/api/v1/docs/${collectionName}/${documentId}`)
        .then(response => response.json())
        .then(data => {
            const collectionContent = document.getElementById('collectionContent');

            const documentDetails = `
                <p class='mb-2'>ID: <span class='text-secondary'>${documentId || 'N/A'}</span></p>
                <p class='mb-2'>Fetched In: <span class='text-secondary'>${data._resp || 'N/A'}</span></p>

                <textarea id='dataTextArea' style='width: 100%; min-height: 200px;' class='bg-main font-mono rounded-md p-4 border-none outline-none active:border-none active:outline-none mb-4'>${JSON.stringify(data.data, null, 2) || ''}</textarea>
                
                <button onclick='validateAndPost("${collectionName}", "${documentId}")' class='btn w-full mt-4 bg-main border-main px-4 py-2 rounded-md text-white hover:bg-primary hover:border-secondary focus:outline-none focus:ring-2 focus:ring-secondary focus:ring-opacity-50'>
                    Update Document
                </button>
            `;
            collectionContent.innerHTML = documentDetails;
        })
        .catch(error => {
            console.error('Error fetching document details:', error);
        });
}

function validateAndPost(collectionName, documentId) {
    const dataTextArea = document.getElementById('dataTextArea');
    const rawJSON = dataTextArea.value.trim();

    try {
        const parsedJSON = JSON.parse(rawJSON);

        fetch(`/api/v1/docs/${collectionName}/${documentId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(parsedJSON)
            })
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(data => {
                new Notify({
                    title: 'Collection Update',
                    text: 'Data has been updated!',
                    backgroundColor: 'var(--accent-primary)',
                    position: 'right bottom',
                    autoclose: true,
                    autotimeout: 3000
                });
            })
            .catch(error => {
                console.error('Error sending POST request:', error);
            });
    } catch (error) {
        new Notify({
            title: 'Collection Update Error',
            text: 'Invalid JSON Syntax!',
            backgroundColor: 'var(--accent-error)',
            position: 'right bottom',
            autoclose: true,
            autotimeout: 3000
        });
    }
}

fetch('/api/v1/docs/collections')
    .then(response => response.json())
    .then(data => {
        const collectionsElement = document.getElementById('collections');

        Object.entries(data).forEach(([dbName, count]) => {
            const anchor = document.createElement('div');
            anchor.className = 'flex relative flex-col p-3 w-full rounded-md border shadow-lg outline-none focus:outline-none focus:border-none card card-border'
            anchor.style = 'cursor: pointer;'

            const div1 = document.createElement('div');
            div1.className = 'flex flex-col items-center pb-2 mb-2 w-full text-white font-sm xl:flex-row';
            div1.textContent = dbName;

            const div2 = document.createElement('div');
            div2.className = 'flex items-center w-full';

            const fileTypePill = document.createElement('div');
            fileTypePill.className = 'px-2 py-1 text-xs leading-none rounded-md border pill border-primary';
            fileTypePill.textContent = 'QDB';

            const recordCountText = document.createElement('div');
            recordCountText.className = 'ml-auto text-xs text-gray-500';
            recordCountText.textContent = `${count} Records`;

            div2.appendChild(fileTypePill);
            div2.appendChild(recordCountText);

            anchor.appendChild(div1);
            anchor.appendChild(div2);

            anchor.addEventListener('click', () => {
                fetchData(dbName, count, page = 1);
            });

            collectionsElement.appendChild(anchor);
        });
    })
    .catch(error => {
        console.error('Error fetching data:', error);
    });

document.addEventListener('DOMContentLoaded', () => {
    const wrapper = document.querySelector('.wrapper');
    const toggleButton = document.getElementById('side-panel-toggle');

    wrapper.classList.remove('side-panel-open');

    toggleButton.addEventListener('click', () => {
        wrapper.classList.toggle('side-panel-open');
    });
});

document.getElementById('prevPage').addEventListener('click', () => {
    if (page > 1) {
        page -= 1;
        const collectionName = document.getElementById('collection_title').innerText;
        const count = parseInt(document.getElementById('record_count').innerText);
        fetchData(collectionName, count, page);
        document.getElementById('current_page').textContent = page;
    }
});

document.getElementById('nextPage').addEventListener('click', () => {
    if (page < totalPages) {
        page += 1;
        const collectionName = document.getElementById('collection_title').innerText;
        const count = parseInt(document.getElementById('record_count').innerText);
        fetchData(collectionName, count, page);
        document.getElementById('current_page').textContent = page;
    }
});