# 🌿 NinjaPx - Your Cozy, Pixel Based Analytics Server 🌸

Hello lovely people! Welcome to NinjaPx, your new favorite minimalistic, privacy-focused pixel tracking tool. 🍄 Whether you're tending to your digital garden or brewing a delightful cup of digital tea, NinjaPx is here to bring a touch of simplicity and charm to your website analytics.

## 🌾 Features at a Glance:
- **Privacy Focused** 🦉: All IP addresses are hashed, keeping your visitors' data as cozy and private as a hidden cottage in the woods.
- **Minimalistic** 🌱: NinjaPx believes in simplicity. With no UI, just straightforward JSON responses, and minimal data collection, we keep things light and breezy.
- **Easy to Use** 🌻: Simply drop the pixel onto your site, and voila! You're set to gather those sweet, sweet stats without disturbing the digital fauna.
- **Extendable** 🛠: While currently in its budding phase, NinjaPx dreams of wild growth. A future where adding more data collection is as easy as planting seeds in your garden.
- **Lightweight Requirements** 🍂: Our SQLite database is as lightweight as a leaf on the wind, ensuring your site stays swift and serene.

## 🍯 What NinjaPx Is Not:
While NinjaPx is a wonderful companion for those looking to gain basic insights into their blog or website's performance, it's not designed for intricate sales funnel tracking or optimizing every step of a user's journey. Think of it more like a lovely, rustic path through the woods rather than a meticulously maintained garden. 🌼

## 📦 Getting Started:
To bring NinjaPx into your digital meadow, follow these simple steps:

1. **Install NinjaPx** - Clone this repository and follow the setup instructions. Make sure you have Go installed on your system. Or use the binary. 
2. **Deploy the Pixel** - Place the NinjaPx pixel on your website. Instructions are included, making it a breeze to integrate.
3. **Enjoy Your Insights** - Access your analytics through the provided API endpoints. Dive into the data like it's a pile of autumn leaves!

### 🌟 Configuration Flags:
To tailor NinjaPx perfectly to yur needs, use these flags when launching:

- `db`: Specifies the path to the SQLite database file. 📁 Default: `ninja.db`
  - Example: `-db="/path/to/your/database/ninja.db"`
- `domain`: The domain name where your application will dance. 🏡 Default: `localhost`
  - Example: `-domain="yourdomain.com"`
- `port`: The port on which NinjaPx will listen for visitors. 🌈 Default: `8080`
  - Example: `-port="8080"`
- `maxMindDB`: The path to your MaxMind GeoLite2 City database. This is what allows NinjaPx to understand where your visitors are wandering from. 🌍 Default: `GeoLite2-City.mmdb`
  - Example: `-maxMindDB="/path/to/GeoLite2-City.mmdb"`

### 🗺 Getting MaxMind GeoLite2:
NinjaPx uses the GeoLite2 database by MaxMind to provide geolocation insights. Here’s how you can acquire it:

1. **Create a MaxMind account**: Head over to [MaxMind's website](https://www.maxmind.com/en/geolite2/signup) and sign up for an account. It's a quick and easy process, much like planting seeds in your garden.
2. **Download GeoLite2-City database**: Once your account is set up, navigate to the GeoLite2 section and download the `GeoLite2-City.mmdb` file. This is the treasure map that NinjaPx uses to locate your visitors.
3. **Specify the database path**: Use the `-maxMindDB` flag to tell NinjaPx where this treasure map is buried in your file system.

### 🚀 Launching NinjaPx:
With your configuration flags in hand, and the MaxMind database nestled in its place, you’re ready to launch NinjaPx. Simply run the following command in your terminal, adjusted with your specific flags:

```
go run main.go -db="path/to/ninja.db" -domain="yourdomain.com" -port="8080" -maxMindDB="path/to/GeoLite2-City.mmdb"
```


## 🌙 API Endpoints:
- `/api/stats/today` - Get a lovely overview of today's website activity.
- `/api/stats/historical` - Wander back through the last 30 days of your site's history
- `/api/stats/historical?days=30&pageURL=/your-page` - You can also define specific URLs you are interested in and how far back you want to go. This could be useful to display a page's stats directly onto that page. 

## 🪴 Extend Your Garden:
While NinjaPx is charmingly straightforward today, it's designed to grow with you. The future of NinjaPx includes automatic data aggregation for custom tables, allowing your analytics to flourish like a well-tended garden.


## 💌 Contributing:
Do you have ideas for how NinjaPx could sprout new features? Perhaps a bug that needs tending? Contributions are always welcome! Just fork this repository, create your branch with improvements, and submit a pull request. Let's grow this garden together! 🌷

## 🍄 License:
NinjaPx is freely available under MIT license. It's like sharing seeds with fellow gardeners – we're all about spreading the joy!

---

Thank you for visiting NinjaPx. May your website analytics be as calming and delightful as a walk through a sun-dappled forest. 🌞
