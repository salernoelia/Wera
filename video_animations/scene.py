from manim import *
from manim.animation.creation import Create

class HitzebedingteTode(MovingCameraScene):
    def construct(self):
        # Data setup for Switzerland and Zurich
        switzerland_data = {
            2000: 302, 2001: 378, 2002: 389, 2003: 1402,
            2004: 244, 2005: 421, 2006: 565, 2007: 213,
            2008: 242, 2009: 337, 2010: 391, 2011: 290,
            2012: 321, 2013: 332, 2014: 165, 2015: 747,
            2016: 291, 2017: 399, 2018: 391, 2019: 336,
            2020: 214, 2021: 87, 2022: 474
        }

        zurich_data = {
            2000: 41, 2001: 57, 2002: 61, 2003: 212,
            2004: 32, 2005: 55, 2006: 83, 2007: 31,
            2008: 33, 2009: 44, 2010: 53, 2011: 41,
            2012: 42, 2013: 54, 2014: 24, 2015: 113,
            2016: 41, 2017: 54, 2018: 63, 2019: 48,
            2020: 26, 2021: 13, 2022: 58
        }
        
        # self.camera.background_color = "#1E1E1E"

        


        # Set up axes
        ax = Axes(
            x_range=[1999, 2023, 1],
            y_range=[0, 1500, 100],
            x_length=10,
            y_length=6,
            tips=False,
            axis_config={"include_numbers": False},
            y_axis_config={"include_numbers": False},
        )
        
        # Add a grid background
        grid = NumberPlane(
            x_range=[1999, 2023, 1],
            y_range=[0, 1500, 100],
            x_length=10,
            y_length=6,
            axis_config={"stroke_color": GREY, "stroke_width": 1},
            background_line_style={"stroke_color": GREY, "stroke_width": 1}
        )
        self.add(grid, ax)

        # Adding custom year labels
        for year in switzerland_data:
            label = Text(str(year), font_size=18, font="Helvetica Neue")
            label.next_to(ax.c2p(year, 0), DOWN)
            label.rotate(-PI / 4)
            self.add(label)

        for cases in range(0, 1500, 100):
            label = Text(str(cases), font_size=24, font="Helvetica Neue")
            label.next_to(ax.c2p(1999, cases), LEFT)
            self.add(label)
            
      

        # Title
        title = Text("Hitzebedingte Todesfälle", font_size=24, font="Helvetica Neue").to_edge(UP)
        self.add(title)

        # Legend
        legend = Text("Quelle: BAFU und BAG", font_size=9, font="Helvetica Neue").to_corner(DOWN + RIGHT)

        self.add(legend)

        # Creating the points and graphs for Zurich and Switzerland data
        zurich_points = [ax.c2p(year, cases) for year, cases in zurich_data.items()]
        zurich_graph = VMobject()
        zurich_graph.set_points_smoothly(zurich_points)
        zurich_graph.set_stroke(BLUE, width=3)

        swiss_points = [ax.c2p(year, cases) for year, cases in switzerland_data.items()]
        swiss_graph = VMobject()
        swiss_graph.set_points_smoothly(swiss_points)
        swiss_graph.set_stroke(YELLOW_B, width=3)

     
        
        # Labels for Zurich and Switzerland graphs
        zurich_label = Text("Zürich", font_size=18, font="Helvetica Neue").next_to(zurich_graph.points[-1], RIGHT)
        swiss_label = Text("Schweiz", font_size=18, font="Helvetica Neue").next_to(swiss_graph.points[-1], RIGHT)
        
        

         # Camera follows the Zurich graph as it's being drawn
        def follow_graph(camera_frame, alpha):
            target_position = zurich_graph.point_from_proportion(alpha)
            camera_frame.move_to(target_position)

        self.camera.frame.set(width=ax.width + 2, height=ax.height + 2)
        self.wait(1)
        self.camera.frame.save_state()
        self.play(self.camera.frame.animate.move_to(zurich_graph.get_start()).set(width=ax.x_length / 5))
        self.play(UpdateFromAlphaFunc(self.camera.frame, follow_graph), Create(zurich_graph, run_time=4))


        self.play(Restore(self.camera.frame))
        
        self.play(FadeIn(zurich_label))
        


        self.wait(4)
       
        # Create the Switzerland graph
        self.play(Create(swiss_graph, run_time=4))
        self.play(FadeIn(swiss_label))
        

        # Highlight peaks above 400 for Switzerland data
        peak_dots_circles = []
        for year, cases in switzerland_data.items():
            if cases > 400:
                circle = Circle(color=RED).scale(0.25).move_to(ax.c2p(year, cases))
                peak_dots_circles.append(circle)
                self.play(GrowFromCenter(circle), run_time=0.5)

        self.wait(1)

        # Remove the peak highlights
        for circle in peak_dots_circles:
            self.play(FadeOut(circle), run_time=0.25)

      

        self.wait(1)




from manim import *

class HeatDissipationInCity(Scene):
    def construct(self):
        # Set the background color
        self.camera.background_color = "#1E1E1E"

        # Create the floor
        floor = Line(start=7*LEFT, end=7*RIGHT, stroke_width=4).shift(2*DOWN)
        self.add(floor)

        # Create buildings with more spacing in between
        buildings = VGroup(
            Rectangle(height=2, width=1.5, fill_color=GRAY, fill_opacity=0.75).next_to(floor, UP, buff=0).shift(5.5*LEFT),
            Rectangle(height=3, width=1.5, fill_color=GRAY, fill_opacity=0.75).next_to(floor, UP, buff=0).shift(3*LEFT),
            Rectangle(height=2.5, width=1.5, fill_color=GRAY, fill_opacity=0.75).next_to(floor, UP, buff=0).shift(0),
            Rectangle(height=3.5, width=1.5, fill_color=GRAY, fill_opacity=0.75).next_to(floor, UP, buff=0).shift(3*RIGHT),
            Rectangle(height=2, width=1.5, fill_color=GRAY, fill_opacity=0.75).next_to(floor, UP, buff=0).shift(5.5*RIGHT)
        )
        self.add(buildings)

        # Create the initial heat line
        heat_line = Line(start=7*LEFT + 2*UP, end=3*LEFT + 2*DOWN, color=YELLOW)
        self.play(Create(heat_line))

        # Heat line vanishes when it hits the floor
        self.play(FadeOut(heat_line))

        # Create radiating rays from buildings
        rays = VGroup()
        for building in buildings:
            for y_offset in [0, 0.5, 1]:  # Add more arrows on the building walls
                ray_left = Arrow(
                    start=building.get_left() + y_offset * UP,
                    end=building.get_left() + 0.5 * LEFT + (0.5 - y_offset) * DOWN,
                    color=ORANGE,
                    buff=0
                ).scale(0.4).set_stroke(width=4)
                ray_right = Arrow(
                    start=building.get_right() + y_offset * UP,
                    end=building.get_right() + 0.5 * RIGHT + (0.5 - y_offset) * DOWN,
                    color=ORANGE,
                    buff=0
                ).scale(0.4).set_stroke(width=4)
                rays.add(ray_left, ray_right)



        self.play(LaggedStartMap(GrowArrow, rays, lag_ratio=0.1))

        self.wait(2)






class MultiNodeConnectedNetwork(Scene):
    def construct(self):
        # Set up a large number of nodes and random edges to create an organic-looking network
        import networkx as nx
        import random

        # Create a random graph using networkx
        G = nx.erdos_renyi_graph(16, 0.3)

        # Remove isolated nodes (including the random one in the top left corner)
        G.remove_nodes_from(list(nx.isolates(G)))

        # Create the network graph in Manim
        graph = Graph(
            list(G.nodes),
            list(G.edges),
            layout="spring",
            layout_scale=5,
            vertex_config={
                "color": WHITE,
                "stroke_color": WHITE,
                "stroke_width": 5,
            },
            edge_config={"color": GREY}
        )

        # Draw the graph
        self.play(Create(graph))

        self.wait(1)

        # Scale down the graph a bit and move it to the left
        self.play(graph.animate.scale(0.8).to_edge(LEFT))

        # Draw a longer arrow in the center
        arrow = Arrow(LEFT, 2 * RIGHT, color=WHITE).move_to(ORIGIN)
        self.play(GrowArrow(arrow))

        self.wait(2)





