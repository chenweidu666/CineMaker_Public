const COS = 'https://cinemaker-1300086205.cos.ap-shanghai.myqcloud.com'

export const drama = {
  id: 13,
  title: 'AI女性Vlog',
  description: '以文艺清新的风格，展现都市女孩的日常生活片段，传递温馨治愈之感。',
  style: '文艺清新',
  episodeCount: 6,
  status: 'production',
  coverImage: `${COS}/images/20260221_200304_c14dfc37.jpeg`,
}

export const episode = {
  id: 83,
  dramaId: 13,
  number: 1,
  title: '周一又是元气满满的一天｜打工人的日常也可以很治愈',
  description: 'AI女孩的周一vlog 闹钟赖床→煎蛋拍照→梧桐路通勤→开会差点走神→独食鸡腿→被苏念姐投喂蛋糕→夕阳下班→番茄鸡蛋面翻车→瑜伽笑场→和白小鸥视频聊天→敷面膜写手账说晚安 普通的一天，但好像也没那么糟。',
  status: 'draft',
}

export const characters = [
  {
    id: 129, name: '姜小卷', outfitName: '睡衣', parentId: 53,
    appearance: '深棕色蓬松短卷发，睡后微乱蓬松，几缕碎发贴在脸颊。圆润鹅蛋脸，白皙肌肤，大而圆的棕色眼睛睡眼惺忪，睫毛浓密，粉色薄唇微微嘟起。穿浅粉色棉质睡衣套装，上衣是圆领长袖，胸口有一只小兔子刺绣。',
    imageUrl: `${COS}/images/20260221_162823_92ed4b55.jpeg`,
  },
  {
    id: 130, name: '姜小卷', outfitName: '居家服', parentId: 53,
    appearance: '深棕色蓬松短卷发，发丝微微外翘带弧度。穿奶白色oversize卫衣，衣摆宽松垂到大腿中部，袖口微微卷起。下身深灰色针织阔腿居家裤。双耳佩戴银色星星耳钉。',
    imageUrl: `${COS}/images/20260221_162853_b1f0dd4e.jpeg`,
  },
  {
    id: 131, name: '姜小卷', outfitName: '通勤装', parentId: 53,
    appearance: '深棕色蓬松短卷发，打理整齐。淡妆精致，淡豆沙色唇釉。穿米色针织开衫搭配白色圆领T恤，下身浅蓝色直筒牛仔裤，白色帆布鞋。斜挎米白色帆布包。',
    imageUrl: `${COS}/images/20260221_162903_3b2ef916.jpeg`,
  },
  {
    id: 98, name: '苏念', outfitName: '职场西装', parentId: 15,
    appearance: '28岁女性，身高165cm，身材纤细挺拔。黑色长发盘成利落的低马尾。五官精致英气，冷玫瑰红唇色。穿黑色修身西装外套搭配白色丝质衬衫，黑色高腰九分西装裤，黑色尖头漆皮高跟鞋。',
    imageUrl: `${COS}/images/20260221_155041_77b6f8d8.jpeg`,
  },
  {
    id: 113, name: '白小鸥', outfitName: '清新吊带裙', parentId: 48,
    appearance: '旧漫风格，乌黑长波浪卷发自然垂落。穿浅绿色吊带裙，清新自然。佩戴珍珠耳坠和精致项链。面带温暖微笑，大眼睛明亮有神。',
    imageUrl: `${COS}/characters/20260221_161725_78e48169.jpeg`,
  },
]

export const scenes = [
  { id: 168, name: '姜小卷的卧室', location: '家中卧室', atmosphere: '温馨安静', description: '姜小卷的单人公寓卧室。浅木色地板，奶白色床品搭配薄荷绿小抱枕，床头柜上有复古台灯和手机支架。', imageUrl: `${COS}/images/20260221_171703_0f8f8033.jpeg` },
  { id: 169, name: '姜小卷的浴室', location: '家中浴室', atmosphere: '清爽干净', description: '白色瓷砖，洗手台上摆着各种瓶瓶罐罐的护肤品，镜子边缘贴着小贴纸。', imageUrl: `${COS}/images/20260221_171503_bcfdbf36.jpeg` },
  { id: 170, name: '姜小卷的厨房', location: '家中厨房', atmosphere: '生活气息', description: '开放式小厨房。白色台面搭配原木橱柜，冰箱门上贴满便签和拍立得照片。', imageUrl: `${COS}/images/20260221_171554_57aada75.jpeg` },
  { id: 171, name: '姜小卷的客厅', location: '家中客厅', atmosphere: '放松日常', description: '小客厅，浅灰色布艺沙发上散落着靠枕和薄毯，原木茶几上有手账本和彩笔。', imageUrl: `${COS}/images/20260221_171517_4eb1515d.jpeg` },
  { id: 172, name: '梧桐路·早晨', location: '梧桐路', atmosphere: '元气清新', description: '两侧高大法国梧桐，人行道浅灰色石板。早晨阳光透过梧桐叶洒下斑驳光影。', imageUrl: `${COS}/images/20260221_171925_127915b5.jpeg` },
  { id: 173, name: '地铁6号线车厢', location: '地铁', atmosphere: '城市日常', description: '早高峰地铁6号线车厢，不锈钢扶手和蓝色座椅，车窗外隧道灯光飞掠。', imageUrl: `${COS}/images/20260221_171847_3b45d5c4.jpeg` },
  { id: 174, name: '姜小卷的会议室', location: '公司会议室', atmosphere: '职场紧凑', description: '白色长桌配灰色转椅，墙上挂着投影幕布，桌面散落着笔记本和水杯。', imageUrl: `${COS}/images/20260221_172318_7fef19f4.jpeg` },
  { id: 175, name: '星辰互动食堂', location: '员工食堂', atmosphere: '热闹平凡', description: '明亮的员工食堂，不锈钢打菜窗口和简洁的四人餐桌。', imageUrl: `${COS}/images/20260221_171834_c28a3b9b.jpeg` },
  { id: 176, name: '姜小卷的工位', location: '公司办公区', atmosphere: '安静专注', description: '开放式办公区，白色桌面上有显示器和台灯，桌角放着小盆栽和马克杯。', imageUrl: `${COS}/images/20260221_172346_0b65229a.jpeg` },
  { id: 177, name: '梧桐路·傍晚', location: '梧桐路', atmosphere: '治愈氛围', description: '傍晚夕阳余晖将天空染成橙粉色渐变，路灯刚刚亮起暖黄色光。', imageUrl: `${COS}/images/20260221_172412_acb4332f.jpeg` },
]

export const storyboards = [
  {
    id: 294, number: 1, title: '闹钟', sceneId: 168, duration: 8,
    characterIds: [129],
    firstFrameDesc: '近景，晨光微弱的卧室，穿浅粉色圆领睡衣（胸口兔子刺绣）的深棕色蓬松短卷发女孩蜷缩在被窝中，几缕碎发贴在脸颊，眉头微皱，手机在床头柜上亮起闹钟界面。',
    middleActionDesc: '穿粉色睡衣的短卷发女孩皱眉从被窝中缓缓伸出一只手摸索床头，粉色薄唇微张嘟囔：「周一……再躺五分钟。」关掉闹钟后手缩回被窝，整个人往被子里缩了缩。',
    lastFrameDesc: '穿粉色睡衣的女孩整个人缩进被窝，只露出蓬松凌乱的短卷发顶，表情从微皱变为放松，手机闹钟已熄灭。',
    composedImage: `${COS}/images/20260221_200304_c14dfc37.jpeg`,
    videoUrl: `${COS}/videos/20260221_222010_04dd926f.mp4`,
    firstFrameImages: [`${COS}/images/20260221_200036_069aef85.jpeg`],
    lastFrameImages: [`${COS}/images/20260221_200304_c14dfc37.jpeg`],
  },
  {
    id: 295, number: 2, title: '洗漱', sceneId: 169, duration: 6,
    characterIds: [129],
    firstFrameDesc: '中景，白色调浴室，穿浅粉色圆领睡衣的短卷发女孩站在洗漱台前面对镜子，棕色大眼睛睡眼惺忪，双手撑在台面上。',
    middleActionDesc: '穿粉色睡衣的短卷发女孩对着镜子拍水乳，眼睛还没完全睁开。她低头拿起一对银色星星耳钉戴上，耳钉在灯光下微微闪光，嘴角微扬说：「嗯，精神了。」',
    lastFrameDesc: '穿粉色睡衣的女孩从撑台面的惺忪状态变为微微挺直身体，双耳戴上闪亮星星耳钉，表情从困倦变为有了几分精神。',
    composedImage: `${COS}/images/20260221_202647_f8ec95d2.jpeg`,
    videoUrl: `${COS}/videos/20260221_222013_84a6c999.mp4`,
    firstFrameImages: [`${COS}/images/20260221_202142_677b1d37.jpeg`],
    lastFrameImages: [`${COS}/images/20260221_202647_f8ec95d2.jpeg`],
  },
  {
    id: 296, number: 3, title: '早餐', sceneId: 170, duration: 8,
    characterIds: [130],
    firstFrameDesc: '中景，温暖的厨房，穿奶白色oversize卫衣（袖口微卷、衣摆垂至大腿中部）配深灰针织阔腿裤的短卷发女孩站在灶台前，一手握锅铲，灶上煎锅里一只完美煎蛋，吐司机旁立着咖啡杯。',
    middleActionDesc: '穿奶白卫衣的短卷发女孩低头看着煎蛋，棕色大眼睛亮了亮，嘴角上扬露出满意笑容。吐司弹出，她转身将咖啡倒进杯子，拿起手机对着摆好的早餐拍照，笑着说：「完美煎蛋，今天运气不错。」',
    lastFrameDesc: '穿奶白卫衣的女孩仍站在灶台前，从专注煎蛋变为举起手机对着餐盘拍照，表情从认真变为开心满足，嘴角上扬。',
    composedImage: `${COS}/images/20260221_204516_a75db895.jpeg`,
    videoUrl: `${COS}/videos/20260221_222028_b1a6b20b.mp4`,
    firstFrameImages: [`${COS}/images/20260221_203314_b55ea68f.jpeg`],
    lastFrameImages: [`${COS}/images/20260221_203925_c9a70126.jpeg`, `${COS}/images/20260221_204516_a75db895.jpeg`],
  },
  {
    id: 297, number: 4, title: '出门', sceneId: 172, duration: 5,
    characterIds: [131],
    firstFrameDesc: '全景，清晨梧桐树荫道，阳光透过树叶洒下斑驳光影，穿米色针织开衫搭白色圆领T恤配浅蓝直筒牛仔裤的短卷发女孩，斜挎有简笔涂鸦图案的米白帆布包，戴着耳机走在路上。',
    middleActionDesc: '穿米色开衫的短卷发女孩戴着耳机走在梧桐树下，微风拂过发丝，她微微侧头感受阳光，嘴角带着淡淡笑意，步伐轻快而悠闲。',
    lastFrameDesc: '穿米色开衫的女孩从画面右侧走到中央位置，步伐从悠闲变为稍加快，似乎想起时间，表情从惬意变为微微紧张。',
    composedImage: `${COS}/images/20260221_205031_c80d76d5.jpeg`,
    videoUrl: `${COS}/videos/20260221_222039_1b1b4ce1.mp4`,
    firstFrameImages: [`${COS}/images/20260221_204828_ed9c0f21.jpeg`],
    lastFrameImages: [`${COS}/images/20260221_205031_c80d76d5.jpeg`],
  },
  {
    id: 298, number: 5, title: '地铁', sceneId: 173, duration: 4,
    characterIds: [131],
    firstFrameDesc: '中景，早高峰地铁车厢内，穿米色针织开衫的短卷发女孩靠着车门站立，戴着耳机，棕色眼睛望向车窗外飞掠的灯光，表情平静放空。',
    middleActionDesc: '穿米色开衫的短卷发女孩靠着车门微微晃动，窗外灯光快速飞掠映在她脸上，她眨了眨眼，神情平静而放空。',
    lastFrameDesc: '穿米色开衫的女孩从站立放空变为微微低头看手机，表情从放空变为平静，车窗外灯光依旧飞掠。',
    composedImage: `${COS}/images/20260221_205946_3c425922.jpeg`,
    videoUrl: `${COS}/videos/20260221_222047_7bb12c88.mp4`,
    firstFrameImages: [`${COS}/images/20260221_204912_19f363af.jpeg`],
    lastFrameImages: [`${COS}/images/20260221_205946_3c425922.jpeg`],
  },
  {
    id: 299, number: 6, title: '开会', sceneId: 174, duration: 8,
    characterIds: [98, 131],
    firstFrameDesc: '中景，明亮的会议室，穿米色针织开衫的短卷发女孩坐在会议桌一侧低头记笔记，对面穿黑色修身西装搭白色丝质衬衫的利落低马尾女性站立发言。',
    middleActionDesc: '穿黑色西装的低马尾女性手指轻点白板，语气沉稳地陈述方案。穿米色开衫的短卷发女孩频频点头，低头偷偷在笔记本上画小人。忽然被点名，她微微一愣，迅速挺直身体：「配色上可以再调一下暖色调比例。」',
    lastFrameDesc: '穿米色开衫的女孩从低头记笔记变为抬头正色发言。穿黑色西装的低马尾女性侧头看向她，冷玫瑰红薄唇微微上扬，从严肃变为认可。',
    composedImage: `${COS}/images/20260221_211105_a341d426.jpeg`,
    videoUrl: `${COS}/videos/20260221_222050_96d0db61.mp4`,
    firstFrameImages: [`${COS}/images/20260221_210900_4e3ea7cb.jpeg`],
    lastFrameImages: [`${COS}/images/20260221_211105_a341d426.jpeg`],
  },
  {
    id: 300, number: 7, title: '午饭', sceneId: 175, duration: 6,
    characterIds: [131],
    firstFrameDesc: '中景，热闹的公司食堂，穿米色针织开衫的短卷发女孩独自坐在靠窗位置，面前餐盘上是鸡腿、青菜和米饭，一只耳朵戴着耳机。',
    middleActionDesc: '穿米色开衫的短卷发女孩一个人安静地吃饭，戴着一只耳机听音乐，嚼着鸡腿时嘴角微微上扬，轻声自言自语：「独食的快乐，你们懂吗？」',
    lastFrameDesc: '穿米色开衫的女孩从端正姿态变为稍稍放松靠后，表情从安静变为满足享受。',
    composedImage: `${COS}/images/20260221_211322_994c4ddd.jpeg`,
    videoUrl: `${COS}/videos/20260221_222057_5e249fac.mp4`,
    firstFrameImages: [`${COS}/images/20260221_210802_b315408e.jpeg`],
    lastFrameImages: [`${COS}/images/20260221_211117_1d5d22ff.jpeg`, `${COS}/images/20260221_211322_994c4ddd.jpeg`],
  },
  {
    id: 301, number: 8, title: '下午茶', sceneId: 176, duration: 6,
    characterIds: [98, 131],
    firstFrameDesc: '近景，办公工位，穿米色针织开衫的短卷发女孩坐在电脑前，一手托腮偷偷用手机看外卖App，桌上放着保温杯和笔记本电脑。',
    middleActionDesc: '穿米色开衫的短卷发女孩正偷偷刷手机，忽然感到肩膀被轻轻拍了一下，桌上不知何时多了一块蛋糕。身后穿黑色修身西装的低马尾女性已转身走远。短卷发女孩看着蛋糕笑着说：「今天份的班，值了。」',
    lastFrameDesc: '穿米色开衫的女孩从托腮看手机变为双手捧着蛋糕，表情从偷闲变为惊喜满足。',
    composedImage: `${COS}/images/20260221_211215_2984e950.jpeg`,
    videoUrl: `${COS}/videos/20260221_222101_0e504566.mp4`,
    firstFrameImages: [`${COS}/images/20260221_210819_43fc8d08.jpeg`],
    lastFrameImages: [`${COS}/images/20260221_211215_2984e950.jpeg`],
  },
  {
    id: 302, number: 9, title: '下班', sceneId: 177, duration: 8,
    characterIds: [131],
    firstFrameDesc: '全景，傍晚梧桐路，天空橙粉色渐变晚霞，穿米色针织开衫搭浅蓝直筒牛仔裤的短卷发女孩斜挎涂鸦帆布包走出写字楼大门。',
    middleActionDesc: '穿米色开衫的短卷发女孩走出大门，深深呼出一口气，双臂向上伸了个大大的懒腰。夕阳余晖洒在脸庞上，她抬头看着晚霞说：「晚霞好好看……拍照。」',
    lastFrameDesc: '穿米色开衫的女孩从走出大门的拘束姿态变为伸完懒腰后的放松状态，面带微笑仰望天空。',
    composedImage: `${COS}/images/20260221_212539_a09a2eec.jpeg`,
    videoUrl: `${COS}/videos/20260221_222105_9d4fd4b4.mp4`,
    firstFrameImages: [`${COS}/images/20260221_211220_e9ce32d6.jpeg`],
    lastFrameImages: [`${COS}/images/20260221_212539_a09a2eec.jpeg`],
  },
  {
    id: 303, number: 10, title: '做饭', sceneId: 170, duration: 8,
    characterIds: [130],
    firstFrameDesc: '中景，厨房灯光温暖，穿奶白色oversize卫衣配深灰针织阔腿裤的短卷发女孩站在灶台前，锅里番茄鸡蛋正在翻炒，旁边煮着面。',
    middleActionDesc: '穿奶白卫衣的短卷发女孩翻炒着番茄鸡蛋，尝了一口汤汁后表情突然僵住——盐放多了。她手忙脚乱地往锅里加水，笑着说：「这已经是厨艺天花板了……啊，盐多了！」',
    lastFrameDesc: '穿奶白卫衣的女孩从自信翻炒变为手忙脚乱加水，表情从得意变为尴尬苦笑。',
    composedImage: `${COS}/images/20260221_212927_e40f807f.jpeg`,
    videoUrl: `${COS}/videos/20260221_222122_1c1232dc.mp4`,
    firstFrameImages: [`${COS}/images/20260221_212729_63964cdf.jpeg`],
    lastFrameImages: [`${COS}/images/20260221_212927_e40f807f.jpeg`],
  },
  {
    id: 304, number: 11, title: '瑜伽', sceneId: 171, duration: 6,
    characterIds: [130],
    firstFrameDesc: '全景，客厅地上铺着瑜伽垫，穿奶白色oversize卫衣配深灰阔腿裤的短卷发女孩正在垫上做拉伸动作，手机立在旁边播放教学视频。',
    middleActionDesc: '穿奶白卫衣的短卷发女孩跟着手机视频做瑜伽拉伸，认真尝试一个高难度姿势，身体晃了晃没忍住笑场，直接坐倒在瑜伽垫上，笑着说：「这谁做得到啊？」',
    lastFrameDesc: '穿奶白卫衣的女孩从认真做拉伸变为笑着瘫坐在瑜伽垫上，姿势完全崩塌。',
    composedImage: `${COS}/images/20260221_213417_d1958f2c.jpeg`,
    videoUrl: `${COS}/videos/20260221_222109_538614fb.mp4`,
    firstFrameImages: [`${COS}/images/20260221_213232_cba7a189.jpeg`],
    lastFrameImages: [`${COS}/images/20260221_213417_d1958f2c.jpeg`],
  },
  {
    id: 305, number: 12, title: '视频聊天', sceneId: 171, duration: 8,
    characterIds: [113, 130],
    firstFrameDesc: '近景，客厅，穿奶白色oversize卫衣的短卷发女孩侧躺在沙发上，手持手机视频通话，屏幕中穿浅绿色吊带裙的波浪卷发女孩正灿烂地笑着。',
    middleActionDesc: '屏幕中穿浅绿色吊带裙的波浪卷发女孩歪头好奇地问：「你今天吃什么？」穿奶白卫衣的短卷发女孩翻了个白眼：「番茄鸡蛋面……别说了。」屏幕中的女孩捧腹大笑：「哈哈哈又是番茄鸡蛋！」',
    lastFrameDesc: '穿奶白卫衣的女孩从无奈变为大笑。屏幕中穿绿色吊带裙的波浪卷发女孩同样笑得前俯后仰。',
    composedImage: `${COS}/images/20260221_215113_f36c2db2.jpeg`,
    videoUrl: `${COS}/videos/20260221_222113_118f52a0.mp4`,
    firstFrameImages: [`${COS}/images/20260221_214803_2906c598.jpeg`],
    lastFrameImages: [`${COS}/images/20260221_215113_f36c2db2.jpeg`],
  },
  {
    id: 306, number: 13, title: '晚安', sceneId: 168, duration: 10,
    characterIds: [129],
    firstFrameDesc: '近景，柔和暖光的卧室，穿浅粉色圆领睡衣（胸口兔子刺绣）的短卷发女孩盘腿坐在床上，脸上贴着面膜，手里拿着笔在手账本上写字。',
    middleActionDesc: '穿粉色睡衣的短卷发女孩一边敷面膜一边在手账本上写写画画。她合上手账，伸了个懒腰，然后轻声说：「普通的一天，但好像也没那么糟。晚安。」说完摘下面膜，把星星耳钉轻轻放在床头柜上，缩进被窝。',
    lastFrameDesc: '穿粉色睡衣的女孩从盘腿写手账变为缩进被窝侧躺，面膜已摘下，表情安然满足，暖光依旧柔和。',
    composedImage: `${COS}/images/20260221_215610_f09feed9.jpeg`,
    videoUrl: `${COS}/videos/20260221_222117_5e980e56.mp4`,
    firstFrameImages: [`${COS}/images/20260221_213842_37d1aab5.jpeg`],
    lastFrameImages: [`${COS}/images/20260221_215516_cf6862f3.jpeg`],
  },
]

export function getCharacterById(id) {
  return characters.find(c => c.id === id)
}

export function getSceneById(id) {
  return scenes.find(s => s.id === id)
}

export function getCharactersForStoryboard(storyboard) {
  return storyboard.characterIds.map(id => getCharacterById(id)).filter(Boolean)
}

export function getSceneForStoryboard(storyboard) {
  return getSceneById(storyboard.sceneId)
}
