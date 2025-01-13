import 'package:flutter/material.dart';
import 'dart:convert';
import 'package:http/http.dart' as http;
import 'videoPage.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';

Future<void> main() async {
  await dotenv.load();
  print(dotenv.env['YOUR_API_KEY']); // Example: access a key from the .env file

  runApp(YoutubeSearchApp());
}

class YoutubeSearchApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'YouTube Search',
      theme: ThemeData(primarySwatch: Colors.red),
      home: YoutubeSearchPage(),
    );
  }
}

class YoutubeSearchPage extends StatefulWidget {
  @override
  _YoutubeSearchPageState createState() => _YoutubeSearchPageState();
}

class _YoutubeSearchPageState extends State<YoutubeSearchPage> {
  final TextEditingController _controller = TextEditingController();
  final String apiKey = dotenv.env['YOUTUBE_API_KEY'] ?? '';

  List<dynamic> _videos = [];

  Future<void> _searchVideos(String query) async {
    final url = Uri.parse(
      'https://www.googleapis.com/youtube/v3/search?part=snippet&type=video&q=$query&key=$apiKey&maxResults=10',
    );

    try {
      final response = await http.get(url);
      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        setState(() {
          _videos = data['items'];
        });
      } else {
        throw Exception('Failed to load data');
      }
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Error fetching data: $e')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('YouTube Search'),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          children: [
            TextField(
              controller: _controller,
              decoration: InputDecoration(
                labelText: 'Search YouTube',
                border: OutlineInputBorder(),
              ),
            ),
            SizedBox(height: 16),
            ElevatedButton(
              onPressed: () {
                if (_controller.text.isNotEmpty) {
                  _searchVideos(_controller.text);
                }
              },
              child: Text('Search'),
            ),
            Expanded(
              child: GridView.builder(
                gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: 2,
                  crossAxisSpacing: 8,
                  mainAxisSpacing: 8,
                ),
                itemCount: _videos.length,
                itemBuilder: (context, index) {
                  final video = _videos[index];
                  final thumbnailUrl =
                      video['snippet']['thumbnails']['high']['url'];
                  final videoId = video['id']['videoId'];
                  return GestureDetector(
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) =>
                              VideoDetailsPage(videoId: videoId),
                        ),
                      );
                    },
                    child: Image.network(thumbnailUrl),
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }
}
