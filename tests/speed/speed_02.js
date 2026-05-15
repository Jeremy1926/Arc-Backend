//const url = "http://account:2525/test/ping";
const url = "http://localhost:2525/test/ping";
setInterval(() => {
    fetch(url, {
        method: "GET",
        headers: { "Content-Type": "application/json", "Authorization": "gqztinbwg43tinzqgu4tizbug43tomnztg44diojwheztmmzyguytiojugu3gmntbg5qtmyjugu2tmnlbgzrgtknjthe2txenrsgzqtmzjvhe3gknbrg44tiyjwmy2gcnjygq4donzvha2genrvgu2dizjuhde3tomzvgyzdmmrtgu2gcnrugrsdinztgq2demzygeq4tmmjxha3gmnduggy3tmmjtgy3gxenzygq4tkmjuge2ggnbugqytoobuhahztgnjxgmytmojwha3tqmzvgu3diyztheztknrzgrrtombvga3genrrgq2dgmztha3dsntggrqtgmzvhe3tanztgrtdkzrwgq3tonjwg4ztonbxguztonrxgrsdonjuge3gkndcgyzdmnrxgy2dsn3bgqztknbumi3domzxg44dmmruhe3genbrgy4dmyzugy3tcnbtguydgmbxgm2tcnzrgu3dgnrumiztcntggu4tkmjxgy2denrygm2tgmbvgq2ggnryg44tmzrvme3wcnjygm3dinjwmm2tgnzxgqytgmrtgq3gkntegvqtmobxg42genryg44tgojvgy2tonbygu4dmnzumm2donzugm2tgmbuha3ggnrwg4ydonrumq2tmntbgy3tkobug43tcnjygztdinjumu3genrygqytknzxg42gmnlbgzrdgnrvgq2tentfgq2tkzrwge3dqnbrgy4tknrvgy2tcnjtg43timzwgy2wcndbgu4dimrugqztinzqgq4dgojvgq2gentbgy2dmmzummztcnbrgu3tkmzxgqztqnbug44tmnbtge2tcnjygq4tmzbvgizginrugq2tmmzvmeztkmzug43dimrwmm3dknddgrrtmmrvme3tg", "X-Arc-Client": "qxvnojzmkarfytwbshlpedgcui" },
    }).catch(error => {
        console.error("Error pinging server:", error);
    })
}, 0.001);